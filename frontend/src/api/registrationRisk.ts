/**
 * 自建顺序点击验证码 API
 */
import { apiClient } from './client'
import type { ClickCaptchaChallenge, ClickCaptchaVerifyResponse } from '@/types'

export async function createClickCaptchaChallenge(): Promise<ClickCaptchaChallenge> {
  const { data } = await apiClient.post<ClickCaptchaChallenge>('/auth/captcha/challenge')
  return data
}

export async function verifyClickCaptcha(
  challengeId: string,
  clicks: string[],
): Promise<ClickCaptchaVerifyResponse> {
  const { data } = await apiClient.post<ClickCaptchaVerifyResponse>('/auth/captcha/verify', {
    challenge_id: challengeId,
    clicks,
  })
  return data
}
