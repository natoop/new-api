package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/model"
	"gorm.io/gorm"
)

type DistributionInvitationCreateInput struct {
	InviteeEmail   string `json:"invitee_email"`
	Level          int    `json:"level"`
	ExpiresAt      int64  `json:"expires_at"`
	IdempotencyKey string `json:"idempotency_key"`
	Remark         string `json:"remark"`
}

type DistributionInvitationAcceptInput struct {
	InvitationNo string `json:"invitation_no"`
}

func CreateDistributionInvitation(inviterUserID int, input DistributionInvitationCreateInput) (*model.DistributionInvitation, error) {
	input.InviteeEmail = strings.TrimSpace(strings.ToLower(input.InviteeEmail))
	input.Remark = strings.TrimSpace(input.Remark)
	if input.InviteeEmail == "" {
		return nil, fmt.Errorf("invitee_email cannot be empty")
	}
	if input.Level < 0 {
		return nil, fmt.Errorf("level cannot be negative")
	}
	key, err := NormalizeIdempotencyKey(input.IdempotencyKey)
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	var invitation model.DistributionInvitation
	err = model.DB.Transaction(func(tx *gorm.DB) error {
		var parentAgent model.DistributionAgent
		if err := distributionLock(tx).Where("user_id = ? AND status = ?", inviterUserID, DistributionStatusEnabled).First(&parentAgent).Error; err != nil {
			return err
		}
		invitationNo := BuildInvitationNo(parentAgent.Id, input.InviteeEmail, key)
		err := tx.Where("invitation_no = ?", invitationNo).First(&invitation).Error
		if err == nil {
			if invitation.InviterUserId != inviterUserID || invitation.InviteeEmail != input.InviteeEmail || invitation.IdempotencyKey != key {
				return fmt.Errorf("idempotency key conflicts")
			}
			return nil
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		invitation = model.DistributionInvitation{
			InvitationNo:   invitationNo,
			IdempotencyKey: key,
			InviteeEmail:   input.InviteeEmail,
			ParentAgentId:  parentAgent.Id,
			Level:          input.Level,
			Status:         DistributionInvitationStatusPending,
			InviterUserId:  inviterUserID,
			ExpiresAt:      input.ExpiresAt,
			Remark:         input.Remark,
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		return tx.Create(&invitation).Error
	})
	if err != nil {
		return nil, err
	}
	return &invitation, nil
}

func ListDistributionInvitations(inviterUserID int) ([]model.DistributionInvitation, error) {
	var invitations []model.DistributionInvitation
	err := model.DB.Where("inviter_user_id = ?", inviterUserID).Order("id desc").Find(&invitations).Error
	return invitations, err
}

func AcceptDistributionInvitation(inviteeUserID int, input DistributionInvitationAcceptInput) (*model.DistributionInvitation, error) {
	invitationNo := strings.TrimSpace(input.InvitationNo)
	if invitationNo == "" {
		return nil, fmt.Errorf("invitation_no cannot be empty")
	}
	now := time.Now().Unix()
	var invitation model.DistributionInvitation
	err := model.DB.Transaction(func(tx *gorm.DB) error {
		if err := distributionLock(tx).Where("invitation_no = ?", invitationNo).First(&invitation).Error; err != nil {
			return err
		}
		if invitation.Status != DistributionInvitationStatusPending {
			return fmt.Errorf("invitation is not pending")
		}
		if invitation.ExpiresAt > 0 && invitation.ExpiresAt < now {
			invitation.Status = DistributionInvitationStatusExpired
			invitation.UpdatedAt = now
			if err := tx.Save(&invitation).Error; err != nil {
				return err
			}
			return fmt.Errorf("invitation expired")
		}
		var existingAgent model.DistributionAgent
		err := tx.Where("user_id = ?", inviteeUserID).First(&existingAgent).Error
		if err == nil {
			return fmt.Errorf("user already has a distribution agent")
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		var user model.User
		if err := tx.Where("id = ?", inviteeUserID).First(&user).Error; err != nil {
			return err
		}
		agentName := strings.TrimSpace(user.Username)
		if agentName == "" {
			agentName = strings.TrimSpace(user.Email)
		}
		if agentName == "" {
			agentName = fmt.Sprintf("user-%d", inviteeUserID)
		}
		agent := model.DistributionAgent{
			UserId:        inviteeUserID,
			Name:          agentName,
			Status:        DistributionStatusEnabled,
			ParentAgentId: invitation.ParentAgentId,
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		if err := tx.Create(&agent).Error; err != nil {
			return err
		}
		if err := ensureDistributionAgentUserRole(tx, inviteeUserID); err != nil {
			return err
		}
		if err := syncDistributionLegacyInvitedCustomers(tx, &agent, inviteeUserID, now); err != nil {
			return err
		}
		invitation.InviteeUserId = inviteeUserID
		invitation.AcceptedAgentId = agent.Id
		invitation.AcceptedAt = now
		invitation.Status = DistributionInvitationStatusAccepted
		invitation.UpdatedAt = now
		if err := tx.Save(&invitation).Error; err != nil {
			return err
		}
		return bindDistributionCustomer(tx, inviteeUserID, invitation.ParentAgentId, DistributionCustomerEventBind, DistributionSourceInvitation, invitation.Id, invitation.InvitationNo, 0, "distribution invitation accepted", now)
	})
	if err != nil {
		return nil, err
	}
	return &invitation, nil
}
