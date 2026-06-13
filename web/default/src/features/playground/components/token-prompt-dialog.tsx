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
import { useState } from 'react'
import { Link } from '@tanstack/react-router'
import { KeyRoundIcon } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { loadPlaygroundToken, savePlaygroundToken } from '../lib/storage'

interface TokenPromptDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSaved: (token: string) => void
}

export function TokenPromptDialog(props: TokenPromptDialogProps) {
  return (
    <Dialog open={props.open} onOpenChange={props.onOpenChange}>
      <DialogContent>
        {/* Mount the form fresh each time the dialog opens so it prefills from
            storage without syncing state in an effect. */}
        {props.open ? (
          <TokenPromptForm
            onOpenChange={props.onOpenChange}
            onSaved={props.onSaved}
          />
        ) : null}
      </DialogContent>
    </Dialog>
  )
}

interface TokenPromptFormProps {
  onOpenChange: (open: boolean) => void
  onSaved: (token: string) => void
}

function TokenPromptForm(props: TokenPromptFormProps) {
  const { t } = useTranslation()
  // Lazy init prefills with any existing token so this doubles as a "change
  // token" form; the parent remounts us on each open to refresh it.
  const [token, setToken] = useState(() => loadPlaygroundToken())

  const trimmed = token.trim()

  const handleSave = () => {
    if (trimmed.length === 0) return
    savePlaygroundToken(trimmed)
    // Replay the pending action BEFORE closing: parents clear their pending
    // state in onOpenChange(false), so onSaved must observe it first.
    props.onSaved(trimmed)
    props.onOpenChange(false)
  }

  return (
    <>
      <DialogHeader>
        <div className='bg-muted text-muted-foreground flex size-9 items-center justify-center rounded-lg'>
          <KeyRoundIcon className='size-4.5' />
        </div>
        <DialogTitle>{t('Use your own API token')}</DialogTitle>
        <DialogDescription>
          {t(
            'The Playground runs on your own token and never spends platform quota directly.'
          )}{' '}
          <Link to='/keys' onClick={() => props.onOpenChange(false)}>
            {t('Create a token')}
          </Link>
        </DialogDescription>
      </DialogHeader>

      <Input
        type='password'
        autoComplete='off'
        autoFocus
        spellCheck={false}
        value={token}
        onChange={(e) => setToken(e.target.value)}
        onKeyDown={(e) => {
          if (e.key === 'Enter') handleSave()
        }}
        placeholder={t('Paste your API token (sk-...)')}
      />

      <DialogFooter>
        <DialogClose render={<Button variant='outline' />}>
          {t('Cancel')}
        </DialogClose>
        <Button onClick={handleSave} disabled={trimmed.length === 0}>
          {t('Save and continue')}
        </Button>
      </DialogFooter>
    </>
  )
}
