import type { MetadataRoute } from 'next'

export default function manifest(): MetadataRoute.Manifest {
  return {
    name: 'Synaply',
    short_name: 'Synaply',
    description: 'Learn new words with spaced repetition. The easiest way to expand your vocabulary.',
    start_url: '/',
    display: 'standalone',
    background_color: '#ffffff', // base light bg
    theme_color: '#2563eb',      // synaply blue
    orientation: 'any',
    icons: [
      {
        src: '/favicon.ico?v=3',
        sizes: 'any',
        type: 'image/x-icon',
      },
      {
        src: '/icon-192.png?v=3',
        sizes: '192x192',
        type: 'image/png',
        purpose: 'maskable',
      },
      {
        src: '/icon-512.png?v=3',
        sizes: '512x512',
        type: 'image/png',
        purpose: 'any',
      },
      {
        src: '/apple-icon.png?v=3',
        sizes: '180x180',
        type: 'image/png',
      },
    ],
  }
}
