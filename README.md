# Руководство по установке и настройке (Traffilk)

В этом руководстве описано, как запустить **основную ноду** (наш Docker-контейнер с дашбордом) и как настроить **остальные ноды** (серверы с Nginx), чтобы они отдавали метрики.

---

## Часть 1: Настройка целевых серверов (Другие ноды)

На каждом сервере, трафик которого мы хотим отслеживать, нужно установить `node_exporter`, чтобы он собирал сетевую статистику, и настроить Nginx для проксирования этих метрик по HTTPS.

### Шаг 1: Установка `node_exporter` (в одну команду)
Выполните этот блок команд на целевом сервере (просто скопируйте и вставьте всё целиком в терминал):

```bash
wget -qO- https://github.com/prometheus/node_exporter/releases/download/v1.6.1/node_exporter-1.6.1.linux-amd64.tar.gz | tar xvz && \
sudo mv node_exporter-1.6.1.linux-amd64/node_exporter /usr/local/bin/ && \
rm -rf node_exporter-1.6.1.linux-amd64 && \
sudo id -u node_exporter &>/dev/null || sudo useradd -rs /bin/false node_exporter && \
echo -e "[Unit]\nDescription=Node Exporter\nAfter=network.target\n\n[Service]\nUser=node_exporter\nGroup=node_exporter\nType=simple\nExecStart=/usr/local/bin/node_exporter\n\n[Install]\nWantedBy=multi-user.target" | sudo tee /etc/systemd/system/node_exporter.service > /dev/null && \
sudo systemctl daemon-reload && \
sudo systemctl enable --now node_exporter
```
*(Теперь экспортер установлен как системный сервис и работает локально на порту `9100`).*

### Шаг 2: Настройка Nginx
Теперь нужно сделать так, чтобы Nginx отдавал эти метрики по HTTPS (например, по пути `/metrics`). 
Откройте конфигурацию вашего сайта в Nginx (например, `/etc/nginx/sites-available/default`) и добавьте следующий блок `location` внутри секции `server { ... }`:

```nginx
server {
    listen 443 ssl;
    server_name vash-server.com;
    
    # ... ваши настройки ssl ...

    # Проксируем запрос на node_exporter
    location /metrics {
        proxy_pass http://localhost:9100/metrics;
        
        # (Опционально) Защита от посторонних. Оставьте только IP главной ноды:
        # allow 123.123.123.123; # IP главной ноды
        # deny all;
    }
}
```

Перезапустите Nginx:
```bash
sudo systemctl restart nginx
```

**Проверка:** Откройте в браузере `https://vash-server.com/metrics`. Вы должны увидеть кучу текста с метриками (в том числе `node_network_receive_bytes_total`).

---

## Часть 2: Настройка Главной Ноды

Главная нода — это сервер, на котором будет крутиться наш интерфейс (Traffilk) и собирать статистику с других нод. 

### Скачивание и Запуск контейнера

1. Убедитесь, что на главной ноде установлены `git`, `docker` и `docker-compose`.
2. Склонируйте проект из вашего Git-репозитория и перейдите в его папку:

```bash
git clone https://github.com/ВАШ_ЛОГИН/ИМЯ_РЕПОЗИТОРИЯ.git traffilk
cd traffilk
```

3. Запустите проект в фоновом режиме:

```bash
docker-compose up -d --build
```

### Добавление нод в Дашборд

1. Откройте в браузере IP главной ноды на порту 8080 (например: `http://main-node-ip:8080`).
2. Нажмите кнопку **"Add Node"**.
3. Введите название сервера (например, "Web Server 1").
4. В поле **Prometheus URL** вставьте ссылку на метрики, которую мы настроили в Части 1 (например: `https://vash-server.com/metrics`).
5. Нажмите **Save**.

**Готово!** Теперь Главная Нода будет раз в час (а также немедленно при запуске) опрашивать URL по HTTPS, парсить скачанные байты и строить красивые графики. База данных надежно сохраняется в папке `./data`.
