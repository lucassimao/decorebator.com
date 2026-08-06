import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import test from 'node:test'
import { fileURLToPath } from 'node:url'
import path from 'node:path'

const webRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')
const repositoryRoot = path.resolve(webRoot, '..')
const locales = ['en', 'de', 'es', 'fr', 'it', 'ja', 'pt']
const accountMenuLabels = {
  en: 'Account',
  de: 'Konto',
  es: 'Cuenta',
  fr: 'Compte',
  it: 'Account',
  ja: 'アカウント',
  pt: 'Conta',
}
const requiredKeys = [
  'metadataTitle',
  'metadataDescription',
  'eyebrow',
  'title',
  'lede',
  'routesLabel',
  'warning.title',
  'warning.body',
  'inApp.eyebrow',
  'inApp.title',
  'inApp.step1',
  'inApp.step2',
  'inApp.step3',
  'inApp.success',
  'email.eyebrow',
  'email.title',
  'email.instructions',
  'email.verification',
  'email.address',
  'email.button',
  'outcome.eyebrow',
  'outcome.title',
  'outcome.profile',
  'outcome.retention',
  'outcome.refund',
  'privacyPolicy',
  'termsOfService',
  'help',
]

function nestedValue(object, dottedKey) {
  return dottedKey.split('.').reduce((value, key) => value?.[key], object)
}

test('every web locale carries complete account-deletion copy', async () => {
  for (const locale of locales) {
    const messages = JSON.parse(
      await readFile(path.join(webRoot, 'messages', `${locale}.json`), 'utf8')
    )

    for (const key of requiredKeys) {
      const value = nestedValue(messages.accountDeletion, key)
      assert.equal(typeof value, 'string', `${locale}: ${key} must be a string`)
      assert.ok(value.trim(), `${locale}: ${key} must not be empty`)
    }
    assert.equal(messages.accountDeletion.email.address, 'privacy@decorebator.com')
    assert.match(messages.accountDeletion.inApp.step2, new RegExp(accountMenuLabels[locale]))
    assert.ok(messages.navigation.deleteAccount)
  }
})

test('the production route preserves the approved deletion safety contract', async () => {
  const page = await readFile(
    path.join(webRoot, 'src/app/[locale]/delete-account/page.tsx'),
    'utf8'
  )
  assert.match(page, /<aside[\s\S]*aria-labelledby="subscription-warning"/)
  assert.match(page, /<h2 id="subscription-warning"/)
  assert.match(page, /mailto:privacy@decorebator\.com/)
  assert.match(page, /href={`\/\$\{locale\}\/privacy`}/)
  assert.match(page, /href={`\/\$\{locale\}\/terms`}/)
})

test('the deletion resource is discoverable and has localized metadata', async () => {
  const [footer, sitemap, layout] = await Promise.all([
    readFile(path.join(webRoot, 'src/components/home/FooterSection.tsx'), 'utf8'),
    readFile(path.join(webRoot, 'src/app/sitemap.ts'), 'utf8'),
    readFile(path.join(webRoot, 'src/app/[locale]/delete-account/layout.tsx'), 'utf8'),
  ])

  assert.ok(footer.includes('href={`/${locale}/delete-account`}'))
  assert.match(footer, /deleteAccount/)
  assert.match(sitemap, /createEntry\(locale, '\/delete-account'/)
  assert.match(layout, /accountDeletion/)
  assert.match(layout, /alternates/)
})

test('mobile locale copy no longer claims an unavailable typed confirmation', async () => {
  const mobileLocales = ['en', 'de', 'es', 'fr', 'it', 'ja', 'pt-BR', 'pt-PT']
  for (const locale of mobileLocales) {
    const messages = JSON.parse(
      await readFile(path.join(repositoryRoot, 'mobile/i18n/locales', `${locale}.json`), 'utf8')
    )
    assert.ok(messages.profile.continueToDeletion)
    assert.ok(messages.profile.finalDeleteWarning)
    assert.ok(messages.profile.deleteNow)
    assert.doesNotMatch(
      messages.profile.accountDeletedMessage,
      /permanent|dauerhaft|definitiv|completamente|完全/i
    )
    assert.equal(messages.profile.typeDeleteToConfirm, undefined)
    assert.ok(messages.settings.account.deleteWarning)
  }
})
