package model

type DistributionAgent struct {
	Id            int     `json:"id"`
	UserId        int     `json:"user_id" gorm:"uniqueIndex;not null"`
	Name          string  `json:"name" gorm:"type:varchar(128);not null"`
	Status        string  `json:"status" gorm:"type:varchar(32);index;not null;default:'enabled'"`
	Balance       float64 `json:"balance" gorm:"type:decimal(10,6);not null;default:0"`
	CommissionBps int     `json:"commission_bps" gorm:"not null;default:0"`
	ParentAgentId int     `json:"parent_agent_id" gorm:"index;not null;default:0"`
	Level         int     `json:"level" gorm:"index;not null;default:2"`
	Contact       string  `json:"contact" gorm:"type:varchar(255);default:''"`
	Remark        string  `json:"remark" gorm:"type:text"`
	CreatedAt     int64   `json:"created_at" gorm:"bigint"`
	UpdatedAt     int64   `json:"updated_at" gorm:"bigint"`
	Username      string  `json:"username,omitempty" gorm:"-"`
	DisplayName   string  `json:"display_name,omitempty" gorm:"-"`
	Email         string  `json:"email,omitempty" gorm:"-"`
}

func (DistributionAgent) TableName() string {
	return "p3_agents"
}

type DistributionPackage struct {
	Id                   int     `json:"id"`
	SubscriptionPlanId   int     `json:"subscription_plan_id" gorm:"index;not null;default:0"`
	SubscriptionTitle    string  `json:"subscription_title" gorm:"type:varchar(128);default:''"`
	SubscriptionSubtitle string  `json:"subscription_subtitle" gorm:"type:varchar(255);default:''"`
	Name                 string  `json:"name" gorm:"type:varchar(128);not null"`
	Sku                  string  `json:"sku" gorm:"type:varchar(64);uniqueIndex;not null"`
	Description          string  `json:"description" gorm:"type:text"`
	Status               string  `json:"status" gorm:"type:varchar(32);index;not null;default:'enabled'"`
	AgentPrice           float64 `json:"agent_price" gorm:"type:decimal(10,6);not null;default:0"`
	RetailPrice          float64 `json:"retail_price" gorm:"type:decimal(10,6);not null;default:0"`
	SecondaryAgentPrice  float64 `json:"secondary_agent_price" gorm:"type:decimal(10,6);not null;default:0"`
	CreditAmount         int     `json:"credit_amount" gorm:"not null;default:0"`
	SortOrder            int     `json:"sort_order" gorm:"index;not null;default:0"`
	CreatedAt            int64   `json:"created_at" gorm:"bigint"`
	UpdatedAt            int64   `json:"updated_at" gorm:"bigint"`
}

func (DistributionPackage) TableName() string {
	return "p3_packages"
}

type DistributionOrder struct {
	Id                         int     `json:"id"`
	OrderNo                    string  `json:"order_no" gorm:"type:varchar(96);uniqueIndex;not null"`
	OrderType                  string  `json:"order_type" gorm:"type:varchar(32);index;not null;default:'inventory'"`
	SubscriptionOrderId        int     `json:"subscription_order_id" gorm:"index;not null;default:0"`
	IdempotencyKey             string  `json:"idempotency_key" gorm:"type:varchar(128);index;not null"`
	AgentId                    int     `json:"agent_id" gorm:"index;not null"`
	AgentUserId                int     `json:"agent_user_id" gorm:"index;not null;default:0"`
	AgentUserName              string  `json:"agent_user_name" gorm:"type:varchar(128);default:''"`
	AgentAgentId               int     `json:"agent_agent_id" gorm:"index;not null;default:0"`
	UserId                     int     `json:"user_id" gorm:"index;not null"`
	BuyUserId                  int     `json:"buy_user_id" gorm:"index;not null;default:0"`
	BuyUserName                string  `json:"buy_user_name" gorm:"type:varchar(128);default:''"`
	BuyerUserId                int     `json:"buyer_user_id" gorm:"index;not null;default:0"`
	BuyerUsername              string  `json:"buyer_username" gorm:"type:varchar(64);default:''"`
	BuyerDisplayName           string  `json:"buyer_display_name" gorm:"type:varchar(128);default:''"`
	BuyerEmail                 string  `json:"buyer_email" gorm:"type:varchar(255);default:''"`
	PackageId                  int     `json:"package_id" gorm:"index;not null"`
	SubscriptionPlanId         int     `json:"subscription_plan_id" gorm:"index;not null;default:0"`
	SubscriptionTitle          string  `json:"subscription_title" gorm:"type:varchar(128);default:''"`
	SubscriptionSubtitle       string  `json:"subscription_subtitle" gorm:"type:varchar(255);default:''"`
	PackageName                string  `json:"package_name" gorm:"type:varchar(128);default:''"`
	PackageSku                 string  `json:"package_sku" gorm:"type:varchar(64);default:''"`
	PackageDescription         string  `json:"package_description" gorm:"type:text"`
	PackageCreditAmount        int     `json:"package_credit_amount" gorm:"not null;default:0"`
	RedeemCodeId               int     `json:"redeem_code_id" gorm:"index;not null;default:0"`
	RedeemCode                 string  `json:"redeem_code" gorm:"type:varchar(96);index;default:''"`
	RedeemCodeOwnerUserId      int     `json:"redeem_code_owner_user_id" gorm:"index;not null;default:0"`
	RedeemCodeOwnerUsername    string  `json:"redeem_code_owner_username" gorm:"type:varchar(64);default:''"`
	RedeemCodeOwnerDisplayName string  `json:"redeem_code_owner_display_name" gorm:"type:varchar(128);default:''"`
	RedeemCodeOwnerEmail       string  `json:"redeem_code_owner_email" gorm:"type:varchar(255);default:''"`
	AgentActiveCode            string  `json:"agent_active_code" gorm:"type:varchar(96);index;default:''"`
	OriginalAmount             float64 `json:"original_amount" gorm:"type:decimal(10,6);not null;default:0"`
	DiscountAmount             float64 `json:"discount_amount" gorm:"type:decimal(10,6);not null;default:0"`
	CreditDeductionAmount      float64 `json:"credit_deduction_amount" gorm:"type:decimal(10,6);not null;default:0"`
	PaidAmount                 float64 `json:"paid_amount" gorm:"type:decimal(10,6);not null;default:0"`
	Quantity                   int     `json:"quantity" gorm:"not null;default:1"`
	UnitAgentPrice             float64 `json:"unit_agent_price" gorm:"type:decimal(10,6);not null;default:0"`
	TotalAgentPrice            float64 `json:"total_agent_price" gorm:"type:decimal(10,6);not null;default:0"`
	RetailPrice                float64 `json:"retail_price" gorm:"type:decimal(10,6);not null;default:0"`
	CommissionBps              int     `json:"commission_bps" gorm:"not null;default:0"`
	CommissionAmount           float64 `json:"commission_amount" gorm:"type:decimal(10,6);not null;default:0"`
	Status                     string  `json:"status" gorm:"type:varchar(32);index;not null;default:'pending'"`
	PaidAt                     int64   `json:"paid_at" gorm:"bigint;default:0"`
	FulfilledAt                int64   `json:"fulfilled_at" gorm:"bigint;default:0"`
	CompletedAt                int64   `json:"completed_at" gorm:"bigint;default:0"`
	CreatedAt                  int64   `json:"created_at" gorm:"bigint"`
	UpdatedAt                  int64   `json:"updated_at" gorm:"bigint"`
}

func (DistributionOrder) TableName() string {
	return "p3_orders"
}

type DistributionInventory struct {
	Id           int     `json:"id"`
	AgentId      int     `json:"agent_id" gorm:"index;not null"`
	OrderId      int     `json:"order_id" gorm:"index;not null"`
	PackageId    int     `json:"package_id" gorm:"index;not null"`
	Status       string  `json:"status" gorm:"type:varchar(32);index;not null;default:'available'"`
	CreditAmount int     `json:"credit_amount" gorm:"not null;default:0"`
	RetailPrice  float64 `json:"retail_price" gorm:"type:decimal(10,6);not null;default:0"`
	InventoryNo  string  `json:"inventory_no" gorm:"type:varchar(96);uniqueIndex;not null"`
	AssignedTo   int     `json:"assigned_to" gorm:"index;not null;default:0"`
	CreatedAt    int64   `json:"created_at" gorm:"bigint"`
	UpdatedAt    int64   `json:"updated_at" gorm:"bigint"`
	Username     string  `json:"username,omitempty" gorm:"-"`
	DisplayName  string  `json:"display_name,omitempty" gorm:"-"`
	Email        string  `json:"email,omitempty" gorm:"-"`
}

func (DistributionInventory) TableName() string {
	return "p3_inventories"
}

type DistributionBalanceAdjustment struct {
	Id             int     `json:"id"`
	ReferenceNo    string  `json:"reference_no" gorm:"type:varchar(96);uniqueIndex;not null"`
	IdempotencyKey string  `json:"idempotency_key" gorm:"type:varchar(128);index;not null"`
	AgentId        int     `json:"agent_id" gorm:"index;not null"`
	Delta          float64 `json:"delta" gorm:"type:decimal(10,6);not null;default:0"`
	BalanceBefore  float64 `json:"balance_before" gorm:"type:decimal(10,6);not null;default:0"`
	BalanceAfter   float64 `json:"balance_after" gorm:"type:decimal(10,6);not null;default:0"`
	Description    string  `json:"description" gorm:"type:text"`
	CreatedAt      int64   `json:"created_at" gorm:"bigint"`
}

func (DistributionBalanceAdjustment) TableName() string {
	return "p3_balance_adjustments"
}

type DistributionBalanceLedger struct {
	Id             int     `json:"id"`
	LedgerNo       string  `json:"ledger_no" gorm:"type:varchar(96);uniqueIndex;not null"`
	IdempotencyKey string  `json:"idempotency_key" gorm:"type:varchar(128);index"`
	AgentId        int     `json:"agent_id" gorm:"index;not null"`
	UserId         int     `json:"user_id" gorm:"index;not null;default:0"`
	OperatorUserId int     `json:"operator_user_id" gorm:"index;not null;default:0"`
	EntryType      string  `json:"entry_type" gorm:"type:varchar(32);index;not null"`
	SourceType     string  `json:"source_type" gorm:"type:varchar(32);index;not null"`
	SourceId       int     `json:"source_id" gorm:"index;not null;default:0"`
	SourceNo       string  `json:"source_no" gorm:"type:varchar(96);index;default:''"`
	Delta          float64 `json:"delta" gorm:"type:decimal(10,6);not null;default:0"`
	BalanceBefore  float64 `json:"balance_before" gorm:"type:decimal(10,6);not null;default:0"`
	BalanceAfter   float64 `json:"balance_after" gorm:"type:decimal(10,6);not null;default:0"`
	Description    string  `json:"description" gorm:"type:text"`
	CreatedAt      int64   `json:"created_at" gorm:"bigint"`
}

func (DistributionBalanceLedger) TableName() string {
	return "p3_balance_ledgers"
}

type DistributionCommissionLog struct {
	Id          int     `json:"id"`
	AgentId     int     `json:"agent_id" gorm:"index;not null"`
	OrderId     int     `json:"order_id" gorm:"index;not null"`
	BaseAmount  float64 `json:"base_amount" gorm:"type:decimal(10,6);not null;default:0"`
	RateBps     int     `json:"rate_bps" gorm:"not null;default:0"`
	Amount      float64 `json:"amount" gorm:"type:decimal(10,6);not null;default:0"`
	Status      string  `json:"status" gorm:"type:varchar(32);index;not null;default:'posted'"`
	Description string  `json:"description" gorm:"type:text"`
	CreatedAt   int64   `json:"created_at" gorm:"bigint"`
}

func (DistributionCommissionLog) TableName() string {
	return "p3_commission_logs"
}

type DistributionProfitLog struct {
	Id             int     `json:"id"`
	ProfitNo       string  `json:"profit_no" gorm:"type:varchar(96);uniqueIndex;not null"`
	IdempotencyKey string  `json:"idempotency_key" gorm:"type:varchar(128);index"`
	AgentId        int     `json:"agent_id" gorm:"index;not null"`
	ChildAgentId   int     `json:"child_agent_id" gorm:"index;not null"`
	OrderId        int     `json:"order_id" gorm:"index;not null"`
	SourceType     string  `json:"source_type" gorm:"type:varchar(32);index;not null"`
	UnitProfit     float64 `json:"unit_profit" gorm:"type:decimal(10,6);not null;default:0"`
	Quantity       int     `json:"quantity" gorm:"not null;default:1"`
	Amount         float64 `json:"amount" gorm:"type:decimal(10,6);not null;default:0"`
	ParentCost     float64 `json:"parent_cost" gorm:"type:decimal(10,6);not null;default:0"`
	SecondaryPrice float64 `json:"secondary_price" gorm:"type:decimal(10,6);not null;default:0"`
	Status         string  `json:"status" gorm:"type:varchar(32);index;not null;default:'posted'"`
	Description    string  `json:"description" gorm:"type:text"`
	CreatedAt      int64   `json:"created_at" gorm:"bigint"`
}

func (DistributionProfitLog) TableName() string {
	return "p3_profit_logs"
}

type DistributionPriceConfig struct {
	Id              int     `json:"id"`
	PackageId       int     `json:"package_id" gorm:"index;not null"`
	TargetType      string  `json:"target_type" gorm:"type:varchar(32);index;not null;default:'level'"`
	CustomerUserId  int     `json:"customer_user_id" gorm:"index;not null;default:0"`
	AgentLevel      int     `json:"agent_level" gorm:"index;not null;default:0"`
	PriceType       string  `json:"price_type" gorm:"type:varchar(32);index;not null;default:'fixed'"`
	PriceValue      float64 `json:"price_value" gorm:"type:decimal(10,6);not null;default:0"`
	Status          string  `json:"status" gorm:"type:varchar(32);index;not null;default:'enabled'"`
	CreatedByUserId int     `json:"created_by_user_id" gorm:"index;not null;default:0"`
	UpdatedByUserId int     `json:"updated_by_user_id" gorm:"index;not null;default:0"`
	Remark          string  `json:"remark" gorm:"type:text"`
	CreatedAt       int64   `json:"created_at" gorm:"bigint"`
	UpdatedAt       int64   `json:"updated_at" gorm:"bigint"`
}

func (DistributionPriceConfig) TableName() string {
	return "p3_agent_price_configs"
}

type DistributionInvitation struct {
	Id              int    `json:"id"`
	InvitationNo    string `json:"invitation_no" gorm:"type:varchar(96);uniqueIndex;not null"`
	IdempotencyKey  string `json:"idempotency_key" gorm:"type:varchar(128);index;not null"`
	InviteeUserId   int    `json:"invitee_user_id" gorm:"index;not null;default:0"`
	InviteeEmail    string `json:"invitee_email" gorm:"type:varchar(255);index;default:''"`
	ParentAgentId   int    `json:"parent_agent_id" gorm:"index;not null"`
	Level           int    `json:"level" gorm:"index;not null;default:0"`
	Status          string `json:"status" gorm:"type:varchar(32);index;not null;default:'pending'"`
	InviterUserId   int    `json:"inviter_user_id" gorm:"index;not null"`
	AcceptedAgentId int    `json:"accepted_agent_id" gorm:"index;not null;default:0"`
	ExpiresAt       int64  `json:"expires_at" gorm:"bigint;index;not null;default:0"`
	AcceptedAt      int64  `json:"accepted_at" gorm:"bigint;not null;default:0"`
	Remark          string `json:"remark" gorm:"type:text"`
	CreatedAt       int64  `json:"created_at" gorm:"bigint"`
	UpdatedAt       int64  `json:"updated_at" gorm:"bigint"`
}

func (DistributionInvitation) TableName() string {
	return "p3_agent_invitations"
}

type DistributionCustomerOwnership struct {
	Id             int    `json:"id"`
	CustomerUserId int    `json:"customer_user_id" gorm:"uniqueIndex;not null"`
	AgentId        int    `json:"agent_id" gorm:"index;not null"`
	SourceType     string `json:"source_type" gorm:"type:varchar(32);index;not null"`
	SourceId       int    `json:"source_id" gorm:"index;not null;default:0"`
	SourceNo       string `json:"source_no" gorm:"type:varchar(96);index;default:''"`
	PromoCodeId    int    `json:"promo_code_id" gorm:"index;not null;default:0"`
	BoundAt        int64  `json:"bound_at" gorm:"bigint;index;not null;default:0"`
	CreatedAt      int64  `json:"created_at" gorm:"bigint"`
	UpdatedAt      int64  `json:"updated_at" gorm:"bigint"`
	Username       string `json:"username,omitempty" gorm:"-"`
	DisplayName    string `json:"display_name,omitempty" gorm:"-"`
	Email          string `json:"email,omitempty" gorm:"-"`
}

func (DistributionCustomerOwnership) TableName() string {
	return "p3_customer_ownerships"
}

type DistributionCustomerAttributionLog struct {
	Id             int    `json:"id"`
	CustomerUserId int    `json:"customer_user_id" gorm:"index;not null"`
	AgentId        int    `json:"agent_id" gorm:"index;not null"`
	SourceType     string `json:"source_type" gorm:"type:varchar(32);index;not null"`
	SourceId       int    `json:"source_id" gorm:"index;not null;default:0"`
	SourceNo       string `json:"source_no" gorm:"type:varchar(96);index;default:''"`
	EventType      string `json:"event_type" gorm:"type:varchar(32);index;not null"`
	PromoCodeId    int    `json:"promo_code_id" gorm:"index;not null;default:0"`
	PromoCode      string `json:"promo_code" gorm:"type:varchar(64);index;default:''"`
	OrderId        int    `json:"order_id" gorm:"index;not null;default:0"`
	Message        string `json:"message" gorm:"type:text"`
	CreatedAt      int64  `json:"created_at" gorm:"bigint"`
}

func (DistributionCustomerAttributionLog) TableName() string {
	return "p3_customer_attribution_logs"
}

type DistributionPromoCode struct {
	Id             int     `json:"id"`
	AgentId        int     `json:"agent_id" gorm:"index;not null"`
	PackageId      int     `json:"package_id" gorm:"index;not null;default:0"`
	Code           string  `json:"code" gorm:"type:varchar(64);uniqueIndex;not null"`
	Status         string  `json:"status" gorm:"type:varchar(32);index;not null;default:'enabled'"`
	DiscountType   string  `json:"discount_type" gorm:"type:varchar(32);not null"`
	DiscountValue  float64 `json:"discount_value" gorm:"type:decimal(10,6);not null;default:0"`
	MaxRedemptions int     `json:"max_redemptions" gorm:"not null;default:0"`
	UsedCount      int     `json:"used_count" gorm:"not null;default:0"`
	StartsAt       int64   `json:"starts_at" gorm:"bigint;index;not null;default:0"`
	ExpiresAt      int64   `json:"expires_at" gorm:"bigint;index;not null;default:0"`
	CreatedAt      int64   `json:"created_at" gorm:"bigint"`
	UpdatedAt      int64   `json:"updated_at" gorm:"bigint"`
	PackageName    string  `json:"package_name,omitempty" gorm:"-"`
}

func (DistributionPromoCode) TableName() string {
	return "p3_promo_codes"
}

type DistributionGiftRule struct {
	Id              int    `json:"id"`
	Name            string `json:"name" gorm:"type:varchar(128);not null"`
	PackageId       int    `json:"package_id" gorm:"index;not null"`
	GiftPackageId   int    `json:"gift_package_id" gorm:"index;not null"`
	TriggerQuantity int    `json:"trigger_quantity" gorm:"not null;default:1"`
	GiftQuantity    int    `json:"gift_quantity" gorm:"not null;default:1"`
	StartsAt        int64  `json:"starts_at" gorm:"bigint;index;not null;default:0"`
	ExpiresAt       int64  `json:"expires_at" gorm:"bigint;index;not null;default:0"`
	Status          string `json:"status" gorm:"type:varchar(32);index;not null;default:'enabled'"`
	CreatedAt       int64  `json:"created_at" gorm:"bigint"`
	UpdatedAt       int64  `json:"updated_at" gorm:"bigint"`
}

func (DistributionGiftRule) TableName() string {
	return "p3_gift_rules"
}

type DistributionOpsDashboardAuthorization struct {
	Id              int    `json:"id"`
	UserId          int    `json:"user_id" gorm:"index;not null"`
	Status          string `json:"status" gorm:"type:varchar(32);index;not null;default:'granted'"`
	GrantedByUserId int    `json:"granted_by_user_id" gorm:"index;not null;default:0"`
	RevokedByUserId int    `json:"revoked_by_user_id" gorm:"index;not null;default:0"`
	GrantedAt       int64  `json:"granted_at" gorm:"bigint;not null;default:0"`
	RevokedAt       int64  `json:"revoked_at" gorm:"bigint;not null;default:0"`
	Remark          string `json:"remark" gorm:"type:text"`
	CreatedAt       int64  `json:"created_at" gorm:"bigint"`
	UpdatedAt       int64  `json:"updated_at" gorm:"bigint"`
}

func (DistributionOpsDashboardAuthorization) TableName() string {
	return "p3_ops_dashboard_authorizations"
}

func distributionMigrationModels() []interface{} {
	return []interface{}{
		&DistributionAgent{},
		&DistributionPackage{},
		&DistributionOrder{},
		&DistributionInventory{},
		&DistributionBalanceAdjustment{},
		&DistributionBalanceLedger{},
		&DistributionCommissionLog{},
		&DistributionProfitLog{},
		&DistributionPriceConfig{},
		&DistributionInvitation{},
		&DistributionCustomerOwnership{},
		&DistributionCustomerAttributionLog{},
		&DistributionPromoCode{},
		&DistributionGiftRule{},
		&DistributionOpsDashboardAuthorization{},
	}
}
