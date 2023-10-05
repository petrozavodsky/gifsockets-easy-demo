# gifsockets-easy-demo

1) Запустить в терминале программу `go run main.go`
2) Открыть в браузере index.html
3) Смотреть вывод терминала

Некоторые параметры конфигурации сервиса можно задать в файле .env

**Для закпуска в качестве сервиса systemd**


1) Подготовка:

клонирование исходников `mkdir /var/www/counter && git clone git@gitlab.lifehacker.ru:devhacker/services/bb-pixel-gif-socket.git /var/www/counter`  

компиляция сервиса `go build -o && chmod u+x counter`

2) Создатьсервис

`touch  /etc/systemd/system/lh-counter.service `

```
[Unit]
Description=lh counter
After=network.target

[Service]
ExecStart=/var/www/counter/counter
WorkingDirectory=/var/www/counter
Restart=always
RestartSec=5s
StandardOutput=append:/var/log/lh-counter/counter-all.log
StandardError=append:/var/log/lh-counter/counter-all.log

[Install]
WantedBy=multi-user.target
```



Выполнить `systemctl daemon-reload && service lh-counter star`


3) Для ротации файлов создать файл /etc/logrotate.d/lh-counter

```
/var/log/lh-counter/*.log {
    rotate 16
    create
    dateext    
    daily
    compress
    delaycompress    
    missingok
    size 500M
}
```



Выполнить `systemctl restart logrotate.service`

Создание  ssl  сертификата `openssl dhparam -out /etc/ssl/dhparam.pem 4096 && certbot install --cert-name c.lifehacker.ru && `

5) Nginx config 


```
server {
    listen 80;
    listen 443 ssl http2;

    server_name c.lifehacker.ru;

    ssl_certificate /etc/letsencrypt/live/c.lifehacker.ru/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/c.lifehacker.ru/privkey.pem;
      
    ssl_protocols TLSv1 TLSv1.1 TLSv1.2;
    ssl_dhparam /etc/ssl/dhparam.pem;
    ssl_prefer_server_ciphers on;
    ssl_ciphers EECDH+CHACHA20:EECDH+AES128:RSA+AES128:EECDH+AES256:RSA+AES256:EECDH+3DES:RSA+3DES:!MD5;
    ssl_session_timeout 1d;
    ssl_session_cache shared:SSL:50m;   
    ssl_stapling on;
    ssl_stapling_verify on;
    resolver 1.1.1.1 8.8.8.8 valid=300s;

    location / {
        return https://lifehacker.ru/;
    }

    location /pixel.gif {
        proxy_pass http://localhost:82/pixel.gif;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```


Выполнить `lh -s /etc/nginx/sites-available/c-lifehacker-ru.conf /etc/nginx/sites-enabled/c-lifehacker-ru.conf && service nginx restart`
