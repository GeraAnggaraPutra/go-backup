# 🗄️ Go Database Backup Tool

A robust and interactive CLI-based utility built with Go, designed to streamline the process of backing up MySQL and PostgreSQL databases. This tool offers a user-friendly experience for selectively archiving database tables into compressed ZIP files.

## ✨ Key Features

- **Interactive Table Selection**: Use a graphical terminal interface (powered by Survey) to pick specific tables you want to back up.
- **Multi-Database Support**: Seamlessly compatible with both MySQL and PostgreSQL engines.
- **Automated ZIP Compression**: Dumps are instantly compressed into ZIP format to save disk space and improve portability.
- **Flexible Configuration**: Supports automated execution via `.env` files or manual input for one-off tasks.
- **Cron Job Integration**: Built-in support for scheduled backups using Cron expressions (e.g., `0 0 * * *` for daily midnight backups).
- **Instant Telegram Notifications**: Receive real-time backup status reports (Success/Failure) directly to your Telegram account or group, complete with file details and timestamps.
- **Real-time Management Dashboard**: A modern web interface to monitor live statistics, manage archives with batch operations, and track server disk capacity via a responsive, auto-refreshing UI.

---

## ⚙️ How It Works

1. **Connectivity**: The app validates the database connection using your provided credentials.
2. **Discovery**: It fetches a real-time list of all available tables within the target database.
3. **Selection**: You interactively select the tables you wish to include in the backup.
4. **Execution**: The tool executes the dump process for the chosen tables.
5. **Archiving**: The resulting data is bundled into a `.zip` file and moved to the designated output folder.
6. **Notification**: Upon completion or if an error occurs mid-process, the bot dispatches a concise report using professional HTML formatting to Telegram.
---

## 🛠️ Usage

### 1. Configuration
The use of a `.env` file is **optional**. If provided in the root directory, the application will pre-load these settings. Otherwise, you can provide all necessary information through the interactive terminal prompts.

Copy file .env.example to .env.
```
cp .env.example .env
```

Configuration Reference:
```env
DB_TYPE=mysql
DB_HOST=127.0.0.1
DB_PORT=3306
DB_USER=root
DB_PASSWORD=password
DB_NAME=database
# Comma-separated list of tables to back up, or empty for all tables
DB_TABLES=table1,table2
OUTPUT_FILE=backup.zip
# Empty string means no schedule, otherwise use cron format
SCHEDULE="0 0 * * *"

# Telegram Notification (Optional)
TELEGRAM_TOKEN=123456789:ABCDefghIJKLmn-Opqrst
TELEGRAM_CHAT_ID=987654321
```

### Method 1: Running Manually (Go)

You can run the tool directly using the Go command.

**Interactive Mode:**
If you run the command without a configured `.env` file (or choose "Manual Input" when prompted), the tool will guide you through the process.

```bash
make backup
```
*Follow the on-screen prompts to select database type, credentials, and tables.*

**Automated Mode:**
If a `.env` file is present, the tool will automatically load the configuration and start the backup.

```bash
make backup
```

### Method 2: Running with Docker
Docker is the recommended way to run the tool in production or environments without Go installed.

**1. Build and Run with Docker Compose**
The project includes a `docker-compose.yml` that handles the setup, including volume mapping for your backup files.
```bash
make docker-build
make docker-backup
```

**2. Check Backups**
Backup files will be generated in the `./backups` directory on your host machine (mapped to `/app/backups` in the container).

**Docker Configuration Notes:**
-   Ensure your `.env` file points to the correct database host. If your database is running on the host machine, use `host.docker.internal` as the `DB_HOST`.

### 2. Dashboard
The web dashboard allows you to manage your backups visually. You can run it with a custom port or use the system default.

**1. Run Dashboard**
```bash
# Run on default port (8080)
make dashboard

# Run on a specific port
make dashboard port=9000

# Docker
make docker-dashboard port=9000 (optional custom port)
```

**2. Accessing the UI**
Once the command is running, open your browser and navigate to:
- Default: http://localhost:8080
- Custom: http://localhost:9000 (or whichever port you specified)

### 3. Output Location
All generated backup files will be found in the following directory:
`./backups/`
