# Traffilk V2

[🇷🇺 Читать на русском](#русская-версия) | [🇬🇧 Read in English](#english-version)

---

## Русская Версия

Легкий, красивый и безопасный дашборд для мониторинга трафика с помощью Prometheus Node Exporter.
Traffilk подключается к вашим серверам с метриками Prometheus и агрегирует входящий/исходящий сетевой трафик за каждый день.

В версии V2 полностью переработан пользовательский интерфейс (в стиле Uptime Kuma) и добавлена встроенная JWT-авторизация.

### Особенности
- **Красивый интерфейс в стиле Uptime Kuma**: Темная тема, закругленные элементы, адаптивный дизайн.
- **Безопасный доступ**: Встроенный экран входа с использованием JWT токенов.
- **Быстрые обновления**: Опрашивает ноды каждую минуту для отображения графиков в реальном времени.
- **Мультиязычность**: Полный перевод интерфейса на русский (RU) и английский (EN) языки.
- **Простое управление**: Добавляйте, изменяйте и удаляйте ноды прямо из браузера.

### 🚀 Установка (Docker)

1. Склонируйте репозиторий:
   ```bash
   git clone https://github.com/alexporteb/Traffilk.git
   cd traffilk
   ```

2. Откройте `docker-compose.yml` и задайте свои логин и пароль в секции `environment`:
   ```yaml
   environment:
     - ADMIN_USER=admin
     - ADMIN_PASS=your_secure_password
     - JWT_SECRET=change-this-secret
   ```

3. Запустите контейнер:
   ```bash
   docker compose up -d
   ```

4. Панель будет доступна по адресу `http://localhost:8080/ui/`.

---

### 🌐 Настройка Nginx / Caddy (Обратный прокси)

Чтобы Traffilk был доступен по красивому домену с HTTPS, используйте один из этих конфигов.
Обратите внимание, что Traffilk работает с подпутями (например, `/traffilk/`), если вам нужно повесить его не на главный домен.

#### Пример Caddyfile:
```caddyfile
# Если ставите на отдельный поддомен:
traffilk.yourdomain.com {
    reverse_proxy localhost:8080
}

# Если ставите в папку существующего домена (например, yourdomain.com/traffilk/):
yourdomain.com {
    route /traffilk/* {
        uri strip_prefix /traffilk
        reverse_proxy localhost:8080
    }
}
```

#### Пример Nginx:
```nginx
server {
    listen 80;
    server_name traffilk.yourdomain.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

---

### 🖥️ Настройка других нод (Серверов)

Чтобы собирать статистику с других серверов, установите на них `node_exporter` одной быстрой командой.

Зайдите на **удалённый** сервер по SSH и выполните:

**Установка:**
```bash
curl -sL https://raw.githubusercontent.com/alexporteb/Traffilk/main/install_node_exporter.sh | bash
```

**Удаление:**
```bash
curl -sL https://raw.githubusercontent.com/alexporteb/Traffilk/main/uninstall_node_exporter.sh | bash
```

После установки метрики будут доступны по адресу `http://<IP_ЭТОГО_СЕРВЕРА>:9100/metrics`.

#### Защита метрик через Caddy / Nginx (Рекомендуется)
Оставлять порт 9100 открытым по HTTP небезопасно. Рекомендуется спрятать его за обратным прокси на удалённой ноде, чтобы получить защищенный HTTPS-доступ. Вы также можете разрешить доступ только для IP-адреса вашего главного сервера Traffilk.

**Пример для Caddy:**
```caddyfile
node1.yourdomain.com {
    # Опционально: Разрешить доступ только главному серверу Traffilk (IP: 123.45.67.89)
    # @blocked not remote_ip 123.45.67.89
    # respond @blocked "Access Denied" 403

    reverse_proxy localhost:9100
}
```

**Пример для Nginx:**
```nginx
server {
    listen 80; # Или 443 с настроенным SSL
    server_name node1.yourdomain.com;

    location / {
        # Опционально: Разрешить доступ только главному серверу Traffilk (IP: 123.45.67.89)
        # allow 123.45.67.89;
        # deny all;

        proxy_pass http://localhost:9100;
    }
}
```

После этого добавьте безопасную ссылку в панели Traffilk: `https://node1.yourdomain.com/metrics`.

*Примечание: При первом добавлении график покажет 0 байт, так как для расчета дневного трафика системе нужно подождать 1 минуту и сделать второй замер.*

<br><br><br>

---

## English Version

A lightweight, beautiful, and secure traffic monitoring dashboard for Prometheus Node Exporter.
Traffilk connects to your existing Prometheus metrics endpoints to fetch and aggregate daily incoming/outgoing network traffic. 

V2 features a completely redesigned user interface (inspired by Uptime Kuma) and built-in JWT authentication.

### Features
- **Beautiful Uptime Kuma-inspired UI**: Dark theme, rounded components, responsive design.
- **Secure Access**: Built-in login screen using JWT cookies.
- **Fast Updates**: Polls nodes every minute to provide real-time traffic deltas.
- **Multi-Language Support**: Fully translated in English (EN) and Russian (RU).
- **Easy Management**: Add, Edit, and Delete nodes seamlessly.

### 🚀 Quick Start (Docker)

1. Clone the repository:
   ```bash
   git clone https://github.com/alexporteb/Traffilk.git
   cd traffilk
   ```

2. Open `docker-compose.yml` and modify the environment variables to set your admin credentials:
   ```yaml
   environment:
     - ADMIN_USER=admin
     - ADMIN_PASS=your_secure_password
     - JWT_SECRET=change-this-secret
   ```

3. Start the container:
   ```bash
   docker compose up -d
   ```

4. Access the dashboard at `http://localhost:8080/ui/`.

---

### 🌐 Reverse Proxy Setup (Caddy / Nginx)

#### Caddy Example:
```caddyfile
# Root domain
traffilk.yourdomain.com {
    reverse_proxy localhost:8080
}

# Subfolder path (/traffilk/)
yourdomain.com {
    route /traffilk/* {
        uri strip_prefix /traffilk
        reverse_proxy localhost:8080
    }
}
```

#### Nginx Example:
```nginx
server {
    listen 80;
    server_name traffilk.yourdomain.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

---

### 🖥️ Setting Up Remote Nodes

For Traffilk to monitor other servers, install `node_exporter` using our quick 1-line script.

SSH into your **remote** server and run:

**Install:**
```bash
curl -sL https://raw.githubusercontent.com/alexporteb/Traffilk/main/install_node_exporter.sh | bash
```

**Uninstall:**
```bash
curl -sL https://raw.githubusercontent.com/alexporteb/Traffilk/main/uninstall_node_exporter.sh | bash
```

Once installed, metrics will be available at `http://<NODE_IP>:9100/metrics`.

#### Securing Metrics with Caddy / Nginx (Recommended)
Exposing port 9100 directly over HTTP is insecure. It's recommended to proxy it through Caddy or Nginx on your remote node to get automatic HTTPS. You can optionally restrict access to only your main Traffilk server's IP address.

**Caddy Example:**
```caddyfile
node1.yourdomain.com {
    # Optional: Allow access ONLY from your main Traffilk server (IP: 123.45.67.89)
    # @blocked not remote_ip 123.45.67.89
    # respond @blocked "Access Denied" 403

    reverse_proxy localhost:9100
}
```

**Nginx Example:**
```nginx
server {
    listen 80; # Or 443 with SSL
    server_name node1.yourdomain.com;

    location / {
        # Optional: Allow access ONLY from your main Traffilk server (IP: 123.45.67.89)
        # allow 123.45.67.89;
        # deny all;

        proxy_pass http://localhost:9100;
    }
}
```

Then, add the node in Traffilk using the secure URL: `https://node1.yourdomain.com/metrics`.

*Note: The chart will initially show 0 bytes for a new node. Traffilk requires at least 1 minute to take a second measurement and calculate the daily traffic delta.*
