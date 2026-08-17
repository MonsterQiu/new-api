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
import { SearchIcon } from '@hugeicons/core-free-icons'
import { HugeiconsIcon } from '@hugeicons/react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import {
  InputGroup,
  InputGroupAddon,
  InputGroupInput,
} from '@/components/ui/input-group'
import { cn } from '@/lib/utils'

import type { DocumentationHeading } from '../types'

type DocumentationNavigationProps = {
  activeHeadingId: string
  headings: DocumentationHeading[]
  onSelect?: () => void
}

export function DocumentationNavigation(props: DocumentationNavigationProps) {
  const { t } = useTranslation()
  const [search, setSearch] = useState('')
  const normalizedSearch = search.trim().toLocaleLowerCase()
  const filteredHeadings = normalizedSearch
    ? props.headings.filter((heading) =>
        heading.label.toLocaleLowerCase().includes(normalizedSearch)
      )
    : props.headings

  return (
    <div className='flex min-h-0 flex-col gap-3'>
      <InputGroup>
        <InputGroupAddon>
          <HugeiconsIcon icon={SearchIcon} strokeWidth={2} />
        </InputGroupAddon>
        <InputGroupInput
          value={search}
          onChange={(event) => setSearch(event.target.value)}
          placeholder={t('Search')}
          aria-label={t('Search')}
        />
      </InputGroup>

      <nav
        className='flex min-h-0 flex-col gap-1 overflow-y-auto'
        aria-label={t('Docs')}
      >
        {filteredHeadings.map((heading) => (
          <a
            key={heading.id}
            href={`#${heading.id}`}
            aria-current={
              props.activeHeadingId === heading.id ? 'location' : undefined
            }
            onClick={props.onSelect}
            className={cn(
              'rounded-lg px-3 py-2 text-sm font-medium transition-colors',
              props.activeHeadingId === heading.id
                ? 'bg-accent text-accent-foreground'
                : 'text-muted-foreground hover:bg-muted hover:text-foreground'
            )}
          >
            {heading.label}
          </a>
        ))}
      </nav>
    </div>
  )
}
