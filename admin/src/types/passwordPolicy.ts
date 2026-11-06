export interface PasswordPolicy {
  version: number
  enabled: boolean
  minLength: number
  requireLetter: boolean
  requireNumber: boolean
  requireUppercase: boolean
  requireLowercase: boolean
  requireSpecial: boolean
  maxAgeDays: number
  expiryWarningDays: number
  historyCount: number
  forceChangeOnCreate: boolean
  forceChangeOnReset: boolean
  allowWeakTemporaryPassword: boolean
  effectiveAt: number
}

export interface PasswordState {
  passwordChangeRequired: boolean
  passwordExpired: boolean
  passwordExpiresAt: number
  passwordExpiryWarning: boolean
}
