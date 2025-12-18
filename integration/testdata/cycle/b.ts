// TODO:[#b] implement storage
// depends-on: #a

export async function persist(data: string) {
  return saveToDB(data)
}

async function saveToDB(data: string) {
  return { stored: true, data }
}
