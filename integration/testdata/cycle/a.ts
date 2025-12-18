// TODO:[#a] wire API layer
// deps: #b

import { persist } from "./b"

export async function handleA(payload: string) {
  await persist(payload)
  return { ok: true }
}
