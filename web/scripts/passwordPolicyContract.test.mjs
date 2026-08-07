import assert from 'node:assert/strict'
import test from 'node:test'

import {
  isPasswordTooLong,
  isPasswordTooShort,
  passwordCodePointLength,
  passwordUtf8ByteLength,
} from '../src/utils/passwordPolicy.ts'

test('counts Unicode code points for the minimum', () => {
  assert.equal(passwordCodePointLength('😀'.repeat(8)), 8)
  assert.equal(isPasswordTooShort('😀'.repeat(7)), true)
  assert.equal(isPasswordTooShort('😀'.repeat(8)), false)
})

test('enforces the bcrypt UTF-8 byte boundary', () => {
  assert.equal(passwordUtf8ByteLength('界'.repeat(24)), 72)
  assert.equal(isPasswordTooLong('界'.repeat(24)), false)
  assert.equal(isPasswordTooLong('界'.repeat(25)), true)
  assert.equal(isPasswordTooLong('a'.repeat(73)), true)
})
