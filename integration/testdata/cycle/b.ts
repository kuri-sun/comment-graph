// TODO:[#b] implement storage
// DEPS: #a

export async function persist(data: string) {
  return saveToDB(data)
}

async function saveToDB(data: string) {
  return { stored: true, data }
}
