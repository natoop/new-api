/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Link } from '@tanstack/react-router'
import {
  ArrowRight,
  BadgeCheck,
  HandCoins,
  Headset,
  Link2,
  PackageOpen,
  ShieldCheck,
  Sparkles,
  UserPlus,
} from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { useStatus } from '@/hooks/use-status'
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from '@/components/ui/accordion'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Progress } from '@/components/ui/progress'
import { CopyButton } from '@/components/copy-button'
import { Main } from '@/components/layout'
import { generateAffiliateLink } from '@/features/wallet/lib'
import { getAgentPromotionProgress } from './api'

const DEFAULT_REQUIRED = 3

export function AgentGuide() {
  const { t } = useTranslation()
  const { status } = useStatus()
  const systemName =
    (status?.system_name as string | undefined) ||
    (status?.data?.system_name as string | undefined) ||
    'GosWith'

  const { data: progressRes, isLoading } = useQuery({
    queryKey: ['agent-promotion-progress'],
    queryFn: getAgentPromotionProgress,
    staleTime: 60 * 1000,
  })
  const progress = progressRes?.data ?? null
  const isAgent = progress?.is_agent === true
  const invited = progress?.invited_count ?? 0
  const paid = progress?.paid_count ?? 0
  const requiredInvites = progress?.required_invites || DEFAULT_REQUIRED
  const requiredPaid = progress?.required_paid || DEFAULT_REQUIRED
  const inviteLink = useMemo(
    () => (progress?.aff_code ? generateAffiliateLink(progress.aff_code) : ''),
    [progress?.aff_code]
  )

  const remainingInvites = Math.max(requiredInvites - invited, 0)
  const remainingPaid = Math.max(requiredPaid - paid, 0)

  let remainingMessage: string
  if (remainingInvites === 0 && remainingPaid === 0) {
    remainingMessage = t(
      'All requirements met — your promotion will complete automatically.'
    )
  } else if (remainingInvites === 0) {
    remainingMessage = t(
      '{{count}} more paying friends and you will be promoted automatically.',
      { count: remainingPaid }
    )
  } else if (remainingPaid === 0) {
    remainingMessage = t(
      'Invite {{count}} more friends to register and you will be promoted automatically.',
      { count: remainingInvites }
    )
  } else {
    remainingMessage = t(
      'Invite {{invites}} more friends, with {{paid}} of them paying, to be promoted automatically.',
      { invites: remainingInvites, paid: remainingPaid }
    )
  }

  // Each benefit card carries one Google-spirit accent on its icon chip.
  const benefits = [
    {
      icon: HandCoins,
      title: t('Commission Rewards'),
      description: t(
        'Earn a commission on every order from your customers, settled automatically.'
      ),
      chipClass:
        'bg-accent-amber/15 text-accent-amber dark:bg-accent-amber/20',
    },
    {
      icon: PackageOpen,
      title: t('Exclusive Inventory Pricing'),
      description: t(
        'Stock up on plans at exclusive wholesale prices below retail.'
      ),
      chipClass: 'bg-accent-blue/12 text-accent-blue dark:bg-accent-blue/20',
    },
    {
      icon: ShieldCheck,
      title: t('Customer Ownership Protection'),
      description: t(
        'Users you invite stay attributed to you for the long term.'
      ),
      chipClass:
        'bg-accent-green/12 text-accent-green dark:bg-accent-green/20',
    },
    {
      icon: Headset,
      title: t('Dedicated Support'),
      description: t(
        'Get priority assistance from our team whenever you need it.'
      ),
      chipClass:
        'bg-accent-coral/12 text-accent-coral dark:bg-accent-coral/20',
    },
  ]

  const steps = [
    {
      icon: Link2,
      title: t('Share your invite link'),
      description: t(
        'Send your exclusive link to friends, communities and social channels.'
      ),
    },
    {
      icon: UserPlus,
      title: t('Friends register and purchase'),
      description: t(
        'Your friends sign up through your link and start using the service.'
      ),
    },
    {
      icon: Sparkles,
      title: t('Automatic promotion'),
      description: t(
        'Once you qualify, you are promoted automatically and commissions are credited instantly.'
      ),
    },
  ]

  const faqs = [
    {
      question: t('How do I get promoted to an agent?'),
      answer: t(
        'Invite {{invites}} friends to register, with {{paid}} of them making a purchase. Promotion happens automatically — no application needed.',
        { invites: requiredInvites, paid: requiredPaid }
      ),
    },
    {
      question: t('How are commissions settled?'),
      answer: t(
        'Commissions are settled automatically per order. Please refer to the actual figures shown in the Agent Center.'
      ),
    },
    {
      question: t('How is customer ownership determined?'),
      answer: t(
        'Users who register through your invite link are attributed to you and remain bound to you long term.'
      ),
    },
    {
      question: t('Does becoming an agent cost anything?'),
      answer: t(
        'No. Meeting the invitation requirements promotes you automatically, free of charge.'
      ),
    },
  ]

  return (
    <Main className='ambient-glow'>
      <div className='min-h-0 flex-1 overflow-y-auto'>
        <div className='mx-auto w-full max-w-5xl px-4 pt-16 pb-20 sm:px-6 sm:pt-24'>
          {/* Hero */}
          <section className='flex flex-col items-center text-center'>
            <Badge variant='outline' className='mb-6 gap-1.5 px-3 py-1'>
              <Sparkles className='h-3.5 w-3.5 text-primary' />
              {isAgent ? t('Agent Partner') : t('Agent Program')}
            </Badge>
            <h1 className='max-w-3xl text-4xl font-bold tracking-tight text-balance sm:text-5xl'>
              {isAgent
                ? t('You are already a {{name}} agent partner', {
                    name: systemName,
                  })
                : t('Become a {{name}} agent partner', { name: systemName })}
            </h1>
            <p className='text-muted-foreground mt-5 max-w-xl text-lg text-balance'>
              {isAgent
                ? t(
                    'Manage your packages, customers and commissions in the Agent Center.'
                  )
                : t(
                    'Share to earn. Invite friends to join, and get promoted to agent automatically.'
                  )}
            </p>
            <div className='mt-8 flex flex-wrap items-center justify-center gap-3'>
              {isAgent ? (
                <Button size='lg' render={<Link to='/agent' />}>
                  {t('Go to Agent Center')}
                  <ArrowRight className='ml-1 h-4 w-4' />
                </Button>
              ) : (
                <CopyButton
                  value={inviteLink}
                  variant='default'
                  size='lg'
                  className='px-6'
                  aria-label={t('Copy Invite Link')}
                >
                  {t('Copy Invite Link')}
                </CopyButton>
              )}
            </div>
          </section>

          {/* Promotion progress (non-agent only) */}
          {!isAgent && !isLoading && (
            <section className='mt-16'>
              <Card className='glass-card overflow-hidden'>
                {/* Google-spirit tri-color accent strip. */}
                <div className='from-accent-blue/60 via-primary/50 to-accent-coral/50 h-1.5 bg-gradient-to-r' />
                <CardContent className='grid gap-8 p-6 sm:p-8 md:grid-cols-2'>
                  <div className='space-y-6'>
                    <div>
                      <h2 className='text-lg font-semibold'>
                        {t('Promotion Progress')}
                      </h2>
                      <p className='text-muted-foreground mt-1 text-sm'>
                        {remainingMessage}
                      </p>
                    </div>
                    <div className='space-y-5'>
                      <div className='space-y-2'>
                        <div className='flex items-center justify-between text-sm'>
                          <span className='font-medium'>
                            {t('Invited friends')}
                          </span>
                          <span className='text-muted-foreground tabular-nums'>
                            {Math.min(invited, requiredInvites)}/
                            {requiredInvites}
                          </span>
                        </div>
                        <Progress
                          value={Math.min(
                            (invited / requiredInvites) * 100,
                            100
                          )}
                        />
                      </div>
                      <div className='space-y-2'>
                        <div className='flex items-center justify-between text-sm'>
                          <span className='font-medium'>
                            {t('Paying friends')}
                          </span>
                          <span className='text-muted-foreground tabular-nums'>
                            {Math.min(paid, requiredPaid)}/{requiredPaid}
                          </span>
                        </div>
                        <Progress
                          value={Math.min((paid / requiredPaid) * 100, 100)}
                        />
                      </div>
                    </div>
                  </div>
                  <div className='flex flex-col justify-center space-y-3'>
                    <p className='text-sm font-medium'>
                      {t('Your Referral Link')}
                    </p>
                    <div className='flex items-center gap-2'>
                      <Input
                        readOnly
                        value={inviteLink}
                        className='font-mono text-xs'
                      />
                      <CopyButton
                        value={inviteLink}
                        variant='outline'
                        tooltip={t('Copy Invite Link')}
                      />
                    </div>
                    <p className='text-muted-foreground text-xs'>
                      {t(
                        'Friends who sign up through this link are attributed to you.'
                      )}
                    </p>
                  </div>
                </CardContent>
              </Card>
            </section>
          )}

          {/* Benefits */}
          <section className='mt-24'>
            <h2 className='text-center text-3xl font-bold tracking-tight'>
              {t('Partner Benefits')}
            </h2>
            <p className='text-muted-foreground mx-auto mt-3 max-w-lg text-center'>
              {t('Everything you get as a {{name}} agent partner.', {
                name: systemName,
              })}
            </p>
            <div className='mt-10 grid gap-4 sm:grid-cols-2 lg:grid-cols-4'>
              {benefits.map((benefit) => (
                <Card
                  key={benefit.title}
                  className='border-border/60 transition-shadow hover:shadow-sm'
                >
                  <CardContent className='p-6'>
                    <div
                      className={`flex h-10 w-10 items-center justify-center rounded-lg ${benefit.chipClass}`}
                    >
                      <benefit.icon className='h-5 w-5' />
                    </div>
                    <h3 className='mt-4 font-semibold'>{benefit.title}</h3>
                    <p className='text-muted-foreground mt-2 text-sm leading-relaxed'>
                      {benefit.description}
                    </p>
                  </CardContent>
                </Card>
              ))}
            </div>
          </section>

          {/* How it works */}
          <section className='mt-24'>
            <h2 className='text-center text-3xl font-bold tracking-tight'>
              {t('How It Works')}
            </h2>
            <div className='mt-12 grid gap-10 sm:grid-cols-3 sm:gap-6'>
              {steps.map((step, index) => (
                <div
                  key={step.title}
                  className='relative flex flex-col items-center text-center'
                >
                  {index < steps.length - 1 && (
                    <div className='via-border absolute top-6 left-[calc(50%+2.5rem)] hidden h-px w-[calc(100%-5rem)] bg-gradient-to-r from-transparent to-transparent sm:block' />
                  )}
                  <div className='bg-primary/10 text-primary flex h-12 w-12 items-center justify-center rounded-full'>
                    <step.icon className='h-5 w-5' />
                  </div>
                  <div className='text-muted-foreground mt-4 text-xs font-medium tracking-widest uppercase'>
                    {t('Step {{number}}', { number: index + 1 })}
                  </div>
                  <h3 className='mt-1 font-semibold'>{step.title}</h3>
                  <p className='text-muted-foreground mt-2 max-w-60 text-sm leading-relaxed'>
                    {step.description}
                  </p>
                </div>
              ))}
            </div>
          </section>

          {/* FAQ */}
          <section className='mx-auto mt-24 max-w-2xl'>
            <h2 className='text-center text-3xl font-bold tracking-tight'>
              {t('Frequently Asked Questions')}
            </h2>
            <Accordion className='mt-8'>
              {faqs.map((faq, index) => (
                <AccordionItem key={faq.question} value={`faq-${index}`}>
                  <AccordionTrigger className='text-left'>
                    {faq.question}
                  </AccordionTrigger>
                  <AccordionContent className='text-muted-foreground'>
                    {faq.answer}
                  </AccordionContent>
                </AccordionItem>
              ))}
            </Accordion>
          </section>

          {/* Final CTA */}
          <section className='mt-24'>
            <Card className='glass-card from-primary/10 bg-gradient-to-br via-transparent to-transparent'>
              <CardContent className='flex flex-col items-center gap-5 p-10 text-center'>
                <BadgeCheck className='text-primary h-8 w-8' />
                <h2 className='text-2xl font-bold tracking-tight'>
                  {t('Want to learn more about partner benefits?')}
                </h2>
                <p className='text-muted-foreground max-w-md'>
                  {t(
                    'Reach out to us to learn more about cooperation opportunities.'
                  )}
                </p>
                <div className='mt-2 flex flex-wrap items-center justify-center gap-3'>
                  <Button render={<Link to='/about' />}>
                    {t('Contact Us')}
                    <ArrowRight className='ml-1 h-4 w-4' />
                  </Button>
                  {isAgent && (
                    <Button variant='outline' render={<Link to='/agent' />}>
                      {t('Go to Agent Center')}
                    </Button>
                  )}
                </div>
              </CardContent>
            </Card>
          </section>
        </div>
      </div>
    </Main>
  )
}
