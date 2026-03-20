import type { Metadata, Viewport } from 'next'
import { JetBrains_Mono } from 'next/font/google'
import { GeistPixelGrid } from 'geist/font/pixel'
import { ThemeProvider } from '@/components/theme-provider'

import './globals.css'

const jetbrainsMono = JetBrains_Mono({
  subsets: ['latin'],
  variable: '--font-mono',
})

export const metadata: Metadata = {
  title: 'Orca AI | Multi-Agent Code Orchestration Platform',
  description:
    'Orca AI is the multi-agent orchestration layer that prevents code conflicts before they happen. Real-time overlap detection, structural risk analysis, and full merge control for teams running concurrent AI coding agents.',
  keywords: [
    'multi-agent orchestration',
    'AI coding agents',
    'code conflict detection',
    'overlap analysis',
    'merge coordination',
    'task decomposition',
    'virtual file system',
    'structural risk analysis',
    'micro-commit protocol',
    'IBM watsonx',
    'developer tools',
    'AI infrastructure',
  ],
  authors: [{ name: 'Orca AI' }],
  creator: 'Orca AI',
  publisher: 'Orca AI',
  robots: {
    index: true,
    follow: true,
    googleBot: {
      index: true,
      follow: true,
      'max-video-preview': -1,
      'max-image-preview': 'large',
      'max-snippet': -1,
    },
  },
  openGraph: {
    type: 'website',
    locale: 'en_US',
    title: 'Orca AI | Multi-Agent Code Orchestration Platform',
    description:
      'The orchestration layer for concurrent AI coding agents. Real-time overlap detection, structural risk analysis, and conflict-free merges.',
    siteName: 'Orca AI',
  },
  twitter: {
    card: 'summary_large_image',
    title: 'Orca AI | Multi-Agent Code Orchestration',
    description:
      'Prevent code conflicts across concurrent AI agents. Real-time overlap detection, risk scoring, and automated merge coordination.',
    creator: '@orcaai',
  },
  category: 'technology',
}

export const viewport: Viewport = {
  themeColor: '#F2F1EA',
  width: 'device-width',
  initialScale: 1,
  maximumScale: 5,
}

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  return (
    <html lang="en" className={`${jetbrainsMono.variable} ${GeistPixelGrid.variable}`} suppressHydrationWarning>
      <body className="font-mono antialiased">
        <ThemeProvider attribute="class" defaultTheme="light" enableSystem={false} disableTransitionOnChange>
          {children}
        </ThemeProvider>
      </body>
    </html>
  )
}
