# message-app
Задание:
1) отправить сообщение через web-интерфейс/restful-endpoint
2) отправить сообщение в rabbitMQ
3) в консумере rabbitMQ получить сообщение и вывести его в консоль


# Инструкция
1. Клонируй к себе: `git clone https://github.com/Shishlyannikovvv/message-app.git`

2. Перейди в папку: `cd message-app`

3. Собери и запусти: `docker compose up --build` 

4. Отправь POST-запрос — `curl -X POST -H "Content-Type: application/json" -d '{"message":"Тестовое сообщение"}' http://localhost:8080/send`  
   Должен ответить: "Message sent to RMQ".

5. Посмотри логи consumer: `docker logs message-app-consumer-1` — увидишь "Received: Тестовое сообщение".

6. RMQ UI для мониторинга: http://localhost:15672 (логин/пароль: guest/guest), проверь очередь "test_queue".