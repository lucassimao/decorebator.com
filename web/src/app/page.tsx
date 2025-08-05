import { redirect } from 'next/navigation'
import { routing } from '../../i18n'

// This page only renders when the app is built statically (output: 'export')
export default function RootPage() {
  redirect(`/${routing.defaultLocale}`)
}
