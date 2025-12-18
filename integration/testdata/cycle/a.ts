// TODO:[#a] wire API layer
// depends-on: #b

import { persist } from "./b"

export async function handleA(payload: string) {
  await persist(payload)
  return { ok: true }
}
