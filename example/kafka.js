const { Kafka } = require("kafkajs")

const kafka = new Kafka({
  clientId: "ette events consumer",
  brokers: ["localhost:29092"],
})

const main = async () => {
  const consumer = kafka.consumer({ groupId: "test-group" })

  await consumer.connect()
  await consumer.subscribe({ topic: "events" })

  await consumer.run({
    eachMessage: async ({ topic, partition, message }) => {
      console.log({
        value: message.value.toString(),
      })
    },
  })
}

main()
