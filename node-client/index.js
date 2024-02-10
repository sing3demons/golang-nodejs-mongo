import Server from './server.js'
import { customAlphabet } from 'nanoid'

// Server.start()

function generateId() {
  const nano = customAlphabet('0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz', 11)
  return nano()
}

for (let i = 0; i < 10; i++) {
  console.log(generateId())
}
