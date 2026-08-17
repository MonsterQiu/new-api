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
import { useMutation } from '@tanstack/react-query'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { Dialog } from '@/components/dialog'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import {
  Field,
  FieldDescription,
  FieldGroup,
  FieldLabel,
} from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { Spinner } from '@/components/ui/spinner'
import { useCopyToClipboard } from '@/hooks/use-copy-to-clipboard'
import { tryPrettyJson } from '@/lib/utils'

import { completeCodexOAuth, startCodexOAuth } from '../../api'

type CodexOAuthDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  onKeyGenerated: (key: string) => void
}

export function CodexOAuthDialog(props: CodexOAuthDialogProps) {
  const { t } = useTranslation()
  const { copyToClipboard } = useCopyToClipboard()
  const [authorizeUrl, setAuthorizeUrl] = useState('')
  const [callbackUrl, setCallbackUrl] = useState('')
  const startMutation = useMutation({ mutationFn: startCodexOAuth })
  const completeMutation = useMutation({ mutationFn: completeCodexOAuth })
  const isStarting = startMutation.isPending
  const isCompleting = completeMutation.isPending
  const isBusy = isStarting || isCompleting

  const handleOpenChange = (nextOpen: boolean) => {
    if (!nextOpen && isBusy) return
    props.onOpenChange(nextOpen)
  }

  const handleStart = async () => {
    setAuthorizeUrl('')
    setCallbackUrl('')
    const authorizationWindow = window.open('about:blank', '_blank')
    if (authorizationWindow) {
      authorizationWindow.opener = null
      authorizationWindow.document.title = t('Opening...')
    }

    try {
      const response = await startMutation.mutateAsync()
      if (!response.success) {
        throw new Error(response.message || t('OAuth start failed'))
      }

      const nextAuthorizeUrl = response.data?.authorize_url?.trim()
      if (!nextAuthorizeUrl) {
        throw new Error(t('OAuth start failed'))
      }

      setAuthorizeUrl(nextAuthorizeUrl)
      if (authorizationWindow && !authorizationWindow.closed) {
        authorizationWindow.location.replace(nextAuthorizeUrl)
        toast.success(t('Opened authorization page'))
      } else {
        toast.warning(t('Please manually copy and open the authorization link'))
      }
    } catch (error) {
      authorizationWindow?.close()
      toast.error(
        error instanceof Error ? error.message : t('OAuth start failed')
      )
    }
  }

  const handleCopyAuthorizeUrl = async () => {
    if (!authorizeUrl) return
    await copyToClipboard(authorizeUrl)
  }

  const handleComplete = async () => {
    const input = callbackUrl.trim()
    if (!input) return

    try {
      const response = await completeMutation.mutateAsync(input)
      if (!response.success) {
        throw new Error(response.message || t('OAuth failed'))
      }

      const key = response.data?.key?.trim()
      if (!key) {
        throw new Error(t('OAuth failed'))
      }

      props.onKeyGenerated(tryPrettyJson(key))
      toast.success(t('Credential generated'))
      props.onOpenChange(false)
    } catch (error) {
      toast.error(error instanceof Error ? error.message : t('OAuth failed'))
    }
  }

  return (
    <Dialog
      open={props.open}
      onOpenChange={handleOpenChange}
      title={t('Codex Authorization')}
      description={t(
        'Generate a Codex OAuth credential and insert it into the channel key field.'
      )}
      contentClassName='sm:max-w-2xl'
      contentHeight='auto'
      footer={
        <>
          <Button
            type='button'
            variant='outline'
            onClick={() => handleOpenChange(false)}
            disabled={isBusy}
          >
            {t('Cancel')}
          </Button>
          <Button
            type='button'
            onClick={handleComplete}
            disabled={!authorizeUrl || !callbackUrl.trim() || isBusy}
          >
            {isCompleting && <Spinner data-icon='inline-start' />}
            {isCompleting ? t('Generating...') : t('Generate credential')}
          </Button>
        </>
      }
    >
      <FieldGroup>
        <Alert>
          <AlertDescription>
            {t(
              '1) Click "Open authorization page" and complete login. 2) Your browser may redirect to localhost (it is OK if the page does not load). 3) Copy the full URL from the address bar and paste it below. 4) Click "Generate credential".'
            )}
          </AlertDescription>
        </Alert>

        <div className='flex flex-wrap gap-2'>
          <Button type='button' onClick={handleStart} disabled={isBusy}>
            {isStarting && <Spinner data-icon='inline-start' />}
            {isStarting ? t('Opening...') : t('Open authorization page')}
          </Button>
          <Button
            type='button'
            variant='outline'
            disabled={!authorizeUrl || isBusy}
            onClick={handleCopyAuthorizeUrl}
          >
            {t('Copy authorization link')}
          </Button>
        </div>

        <Field data-disabled={!authorizeUrl || isBusy || undefined}>
          <FieldLabel htmlFor='codex-oauth-callback-url'>
            {t('Callback URL')}
          </FieldLabel>
          <Input
            id='codex-oauth-callback-url'
            value={callbackUrl}
            onChange={(event) => setCallbackUrl(event.target.value)}
            placeholder={t(
              'Paste the full callback URL (includes code & state)'
            )}
            autoComplete='off'
            spellCheck={false}
            disabled={!authorizeUrl || isBusy}
          />
          <FieldDescription>
            {t(
              'Tip: The generated key is a JSON credential including access_token / refresh_token / account_id.'
            )}
          </FieldDescription>
        </Field>
      </FieldGroup>
    </Dialog>
  )
}
