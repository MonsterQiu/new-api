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
export const REQUIRED_LEGAL_LINKS = [
  {
    key: 'terms-of-service',
    path: '/terms-of-service',
    titleKey: 'Terms of Service',
  },
  {
    key: 'usage-policy',
    path: '/usage-policy',
    titleKey: 'Usage Policy',
  },
  {
    key: 'supported-countries',
    path: '/supported-countries',
    titleKey: 'Supported Countries and Regions',
  },
  {
    key: 'service-terms',
    path: '/service-terms',
    titleKey: 'Service-specific Terms',
  },
] as const

export type RequiredLegalDocumentKey =
  (typeof REQUIRED_LEGAL_LINKS)[number]['key']
