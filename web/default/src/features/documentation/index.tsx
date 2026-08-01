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
import { Menu01Icon } from '@hugeicons/core-free-icons'
import { HugeiconsIcon } from '@hugeicons/react'
import { useQuery } from '@tanstack/react-query'
import { useEffect, useMemo, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { ErrorState } from '@/components/error-state'
import { PublicLayout } from '@/components/layout'
import { RichContent } from '@/components/rich-content'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet'
import { Skeleton } from '@/components/ui/skeleton'
import { useStatus } from '@/hooks/use-status'
import { useSystemConfig } from '@/hooks/use-system-config'

import { fetchDocumentationMarkdown } from './api'
import { DocumentationNavigation } from './components/documentation-navigation'
import type { DocumentationHeading } from './types'

const headingMarkdownPattern = /^##\s+(.+)$/gm

function getServerAddress(status: unknown): string {
  if (status && typeof status === 'object') {
    const candidate = (status as Record<string, unknown>).server_address
    if (typeof candidate === 'string' && candidate.trim()) {
      return candidate.trim().replace(/\/+$/, '')
    }
  }

  return window.location.origin
}

export function Documentation() {
  const { t } = useTranslation()
  const contentRef = useRef<HTMLDivElement>(null)
  const [activeHeadingId, setActiveHeadingId] = useState('')
  const [mobileNavigationOpen, setMobileNavigationOpen] = useState(false)
  const { status } = useStatus()
  const { systemName } = useSystemConfig()
  const documentationQuery = useQuery({
    queryKey: ['documentation', 'markdown'],
    queryFn: fetchDocumentationMarkdown,
    staleTime: Number.POSITIVE_INFINITY,
  })

  const headings = useMemo<DocumentationHeading[]>(() => {
    const markdown = documentationQuery.data ?? ''
    return [...markdown.matchAll(headingMarkdownPattern)].map(
      (match, index) => ({
        id: `documentation-section-${index + 1}`,
        label: match[1].replaceAll(/[*_`~]/g, '').trim(),
      })
    )
  }, [documentationQuery.data])

  const renderedMarkdown = useMemo(() => {
    const apiRoot = getServerAddress(status)
    return (documentationQuery.data ?? '')
      .replaceAll('{{SYSTEM_NAME}}', systemName)
      .replaceAll('{{API_ROOT_URL}}', apiRoot)
      .replaceAll('{{API_BASE_URL}}', `${apiRoot}/v1`)
  }, [documentationQuery.data, status, systemName])

  useEffect(() => {
    const renderedHeadings = contentRef.current?.querySelectorAll('h2')
    if (!renderedHeadings || renderedHeadings.length === 0) return

    renderedHeadings.forEach((heading, index) => {
      const navigationHeading = headings[index]
      if (!navigationHeading) return
      heading.id = navigationHeading.id
      heading.classList.add('scroll-mt-24')
    })

    const observer = new IntersectionObserver(
      (entries) => {
        const activeEntry = entries.find((entry) => entry.isIntersecting)
        if (activeEntry) setActiveHeadingId(activeEntry.target.id)
      },
      { rootMargin: '-96px 0px -68% 0px', threshold: 0.01 }
    )

    renderedHeadings.forEach((heading) => observer.observe(heading))
    return () => observer.disconnect()
  }, [headings, renderedMarkdown])

  if (documentationQuery.isLoading) {
    return (
      <PublicLayout showMainContainer={false}>
        <div className='mx-auto grid min-h-svh max-w-7xl gap-8 px-4 pt-24 pb-12 lg:grid-cols-[16rem_minmax(0,1fr)] lg:px-6'>
          <Skeleton className='hidden h-[32rem] lg:block' />
          <div className='flex flex-col gap-4'>
            <Skeleton className='h-12 w-2/3' />
            <Skeleton className='h-5 w-full' />
            <Skeleton className='h-5 w-5/6' />
            <Skeleton className='h-80 w-full' />
          </div>
        </div>
      </PublicLayout>
    )
  }

  if (documentationQuery.isError) {
    return (
      <PublicLayout>
        <ErrorState
          title={t('Failed to load')}
          onRetry={() => documentationQuery.refetch()}
          className='mx-auto mt-12 max-w-3xl'
        />
      </PublicLayout>
    )
  }

  return (
    <PublicLayout showMainContainer={false}>
      <div className='mx-auto grid min-h-svh max-w-7xl gap-8 px-4 pt-24 pb-16 lg:grid-cols-[16rem_minmax(0,1fr)] lg:px-6'>
        <aside className='hidden lg:block'>
          <Card className='sticky top-20 max-h-[calc(100svh-6rem)]'>
            <CardHeader>
              <CardTitle>{t('Usage guide')}</CardTitle>
            </CardHeader>
            <CardContent className='min-h-0'>
              <DocumentationNavigation
                activeHeadingId={activeHeadingId}
                headings={headings}
              />
            </CardContent>
          </Card>
        </aside>

        <main className='min-w-0'>
          <div className='mb-4 lg:hidden'>
            <Button
              variant='outline'
              onClick={() => setMobileNavigationOpen(true)}
            >
              <HugeiconsIcon
                icon={Menu01Icon}
                strokeWidth={2}
                data-icon='inline-start'
              />
              {t('Usage guide')}
            </Button>
          </div>

          <Card>
            <CardContent className='px-5 py-3 sm:px-8 sm:py-6 lg:px-12 lg:py-8'>
              <div ref={contentRef}>
                <RichContent
                  mode='markdown'
                  content={renderedMarkdown}
                  className='max-w-none [&_h1]:text-3xl [&_h1]:tracking-tight sm:[&_h1]:text-4xl [&_h2]:border-b [&_h2]:pb-3 [&_img]:border'
                />
              </div>
            </CardContent>
          </Card>
        </main>
      </div>

      <Sheet open={mobileNavigationOpen} onOpenChange={setMobileNavigationOpen}>
        <SheetContent side='left' className='w-[min(88vw,20rem)]'>
          <SheetHeader>
            <SheetTitle>{t('Docs')}</SheetTitle>
            <SheetDescription>{t('Usage guide')}</SheetDescription>
          </SheetHeader>
          <div className='min-h-0 overflow-y-auto px-4 pb-4'>
            <DocumentationNavigation
              activeHeadingId={activeHeadingId}
              headings={headings}
              onSelect={() => setMobileNavigationOpen(false)}
            />
          </div>
        </SheetContent>
      </Sheet>
    </PublicLayout>
  )
}
