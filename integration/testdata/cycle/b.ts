// TODO: implement storage
// @todo-id b
// @todo-deps a

export async function persist(data: string) {
  return saveToDB(data)
}

async function saveToDB(data: string) {
  return { stored: true, data }
}
