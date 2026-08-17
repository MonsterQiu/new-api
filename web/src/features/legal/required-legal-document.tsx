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
import { useTranslation } from 'react-i18next'

import { PublicLayout } from '@/components/layout'
import { RichContent } from '@/components/rich-content'

import { getRequiredLegalDocument } from './required-legal-documents'
import type { RequiredLegalDocumentKey } from './required-legal-links'

interface RequiredLegalDocumentProps {
  documentKey: RequiredLegalDocumentKey
}

export function RequiredLegalDocument({
  documentKey,
}: RequiredLegalDocumentProps) {
  const { t } = useTranslation()
  const document = getRequiredLegalDocument(documentKey)

  if (!document) {
    return null
  }

  return (
    <PublicLayout>
      <article className='mx-auto max-w-4xl space-y-6 py-12'>
        <h1 className='text-3xl font-semibold tracking-tight'>
          {t(document.titleKey)}
        </h1>
        <RichContent
          mode='markdown'
          content={document.content}
          className='prose-neutral dark:prose-invert max-w-none'
        />
      </article>
    </PublicLayout>
  )
}
