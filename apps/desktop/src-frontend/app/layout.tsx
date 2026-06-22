import type { Metadata, Viewport } from 'next'
import { Inter } from 'next/font/google'
import './globals.css'
import '@/lib/polyfill'

const inter = Inter({ subsets: ['latin'] })

export const metadata: Metadata = {
  title: 'vexa — Complete SSH Manager',
  description: 'Self-hosted SSH access management (desktop roadmap)',
  keywords: ['SSH', 'terminal', 'SFTP', 'remote', 'devops'],
  authors: [{ name: 'vexa Team' }],
  applicationName: 'vexa',
}

export const viewport: Viewport = {
  width: 'device-width',
  initialScale: 1,
  themeColor: '#0f0f0f',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en" className="dark">
      <body className={`${inter.className} antialiased`}>
        {children}
      </body>
    </html>
  )
}
