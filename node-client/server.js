import express from 'express'
import axios from 'axios'
import fs from 'fs'

class Server {
  static async start() {
    const ids = fs.readFileSync('id.txt', 'utf8')

    const app = express()
    const port = 3000

    app.use(express.json())
    app.use(express.urlencoded({ extended: true }))

    app.get('/', async (req, res) => {
      const start = Date.now()

      const products = []
      const productId = ids.split(',')
      console.log(productId.length)
      for (const pId of productId) {
        products.push(getProducts(pId))
      }

      const response = await Promise.all(products)

      res.json({
        duration: (new Date() - start).toFixed(2) + 'ms',
        products: response.slice(0, 1000),
        status: 'success',
        total: response.length,
      })
    })

    const server = app.listen(port, () => console.log('Server is running on port 3000'))

    process.on('SIGTERM', () => {
      console.log('SIGTERM signal received: closing HTTP server')
      server.close(() => process.exit(0))
    })
    process.on('SIGINT', () => {
      console.log('SIGINT signal received: closing HTTP server')
      server.close(() => process.exit(0))
    })
  }
}

export default Server

async function getProducts(id) {
  const { data } = await axios.get('http://localhost:8080/products/' + id, {
    headers: {
      'Content-Type': 'application/json',
    },
    timeout: 30000,
  })
  return data
}
