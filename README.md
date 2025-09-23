Не смог подключиться к сокету выдает ошибку на сайте :
Cross-Origin Request Blocked: The Same Origin Policy disallows reading the remote resource at http://localhost:5200/socket.io/?EIO=4&transport=polling&t=ee28q8hp. (Reason: CORS header ‘Access-Control-Allow-Origin’ missing). Status code: 200.
Попытался запустить в окружении, но тоже не получилось.
heatingeventservicego:
    build:
      context: ../HeatingEventServiceGo
      dockerfile: Dockerfile
    container_name: heatingeventservicego
    depends_on:
      - postgres
    networks:
      - network
    ports:
      - 5200:5200
    environment:
      PORT: 5200
      DB_HOST: postgres
      DB_USER: postgres
      DB_PASSWORD: postgres
      DB_NAME: messages
    restart: always
          
Кажется я невижу маленькую ошибочку(мб не поменял какой-то порт или не туда подключаюсь) или мне надо во фронте что-то посмотреть и изменить.