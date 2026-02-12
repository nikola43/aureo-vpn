# 👑 Documento Comercial

### La VPN Descentralizada con Recompensas en Criptomonedas

*Privacidad real. Descentralizada. Recompensada.*

---

## 🎯 1. Resumen Ejecutivo

**Aureo VPN** es una plataforma de red privada virtual (VPN) descentralizada que combina privacidad de nivel militar con un modelo economico innovador basado en blockchain. A diferencia de las VPN tradicionales que dependen de infraestructura centralizada, Aureo permite que cualquier persona opere un nodo VPN y genere ingresos pasivos en criptomonedas por compartir ancho de banda.

{% hint style="info" %}
**Propuesta de valor:**
- 👤 Para usuarios finales: privacidad real, sin registros, cifrado de ultima generacion
- 📡 Para operadores de nodos: ingresos pasivos en ETH, BTC o LTC por cada GB servido
- 💼 Para empresas: infraestructura VPN autogestionable con total control y transparencia
{% endhint %}

---

## ⚠️ 2. El Problema

| Problema | VPNs Tradicionales | Aureo VPN |
|----------|-------------------|-----------|
| Centralizacion | Un solo proveedor controla todos los servidores | Red distribuida de nodos independientes |
| Confianza | "No-logs" pero imposible de verificar | Codigo abierto, auditable por la comunidad |
| Punto unico de fallo | Si el proveedor cae, todos los usuarios pierden acceso | Red P2P sin dependencia central |
| Modelo economico | Solo el proveedor genera ingresos | Cualquiera puede operar nodos y ganar |
| Transparencia | Codigo cerrado, politicas opacas | Open source, sistema de reputacion publico |

---

## ⚙️ 3. Como Funciona

### 🏗️ 3.1 Arquitectura del Sistema

Aureo VPN esta compuesto por cuatro componentes principales:

```
+-------------------+       +-------------------+       +-------------------+
|                   |       |                   |       |                   |
|   Cliente Aureo   | <---> |   API Gateway     | <---> |  Control Server   |
|   (Desktop App)   |  API  |   (REST + Auth)   |       |  (Orquestacion)   |
|                   |       |                   |       |                   |
+-------------------+       +-------------------+       +-------------------+
        |                                                        |
        |  WireGuard                                  Heartbeat  |
        |  Tunnel                                     cada 30s   |
        v                                                        v
+-------------------+       +-------------------+       +-------------------+
|                   |       |                   |       |                   |
|   Nodo VPN #1     | <---> |   Nodo VPN #2     | <---> |   Nodo VPN #N     |
|   (Operador A)    |  P2P  |   (Operador B)    |  P2P  |   (Operador N)    |
|                   |       |                   |       |                   |
+-------------------+       +-------------------+       +-------------------+
        |                           |                           |
        v                           v                           v
   Blockchain               Blockchain                  Blockchain
   (Pagos ETH/BTC/LTC)     (Pagos ETH/BTC/LTC)        (Pagos ETH/BTC/LTC)
```

🔑 **API Gateway** - Punto de entrada para todas las solicitudes de los clientes. Maneja autenticacion JWT, gestion de usuarios, descubrimiento de nodos y creacion de sesiones VPN. Base de datos SQLite embebida (cero configuracion).

🎛️ **Control Server** - Orquesta la red de nodos. Ejecuta health checks cada minuto, balanceo de carga cada 30 segundos y limpieza de recursos cada hora. Monitorea el estado de todos los nodos activos.

📡 **Nodos VPN** - Los servidores que manejan las conexiones VPN reales. Soportan WireGuard (puerto 51820/UDP) y OpenVPN (puerto 1194/UDP). Envian heartbeats al Control Server cada 30 segundos con metricas de rendimiento. Cualquier persona puede operar un nodo.

🌐 **Red P2P** - Basada en libp2p con Kademlia DHT para descubrimiento descentralizado de nodos, Gossipsub para mensajeria y mDNS para descubrimiento en red local. Elimina puntos unicos de fallo.

### 👤 3.2 Flujo de Conexion del Usuario

```
1. REGISTRO          El usuario crea su cuenta (email + contrasena)
       |
       v
2. AUTENTICACION     Recibe tokens JWT (access: 15min, refresh: 7 dias)
       |
       v
3. DESCUBRIMIENTO    Consulta nodos disponibles filtrados por pais, protocolo o carga
       |
       v
4. SELECCION         Elige un nodo (manual, Quick Connect, Secure Core, P2P o aleatorio)
       |
       v
5. INTERCAMBIO       Genera par de claves WireGuard y registra clave publica en el nodo
   DE CLAVES
       |
       v
6. CONFIGURACION     Recibe: clave publica del servidor, endpoint, IP del tunel, DNS
       |
       v
7. CONEXION          WireGuard levanta interfaz virtual y establece tunel cifrado
       |
       v
8. TRAFICO           Todo el trafico pasa por el tunel con cifrado ChaCha20-Poly1305
   PROTEGIDO
       |
       v
9. MONITOREO         Stats en tiempo real: velocidad, bytes transferidos, latencia
       |
       v
10. DESCONEXION      Se cierra el tunel, se finalizan las ganancias del operador
```

### 🖥️ 3.3 Aplicacion de Escritorio (Cliente)

El cliente de Aureo VPN es una aplicacion nativa de escritorio construida con **Wails** (Go + Web), disponible para macOS, Windows y Linux.

**Pantalla de inicio de sesion:**
- Login / Registro con email y contrasena
- Configuracion de servidor API

**Dashboard principal:**
- Panel de estado de conexion con IP actual y estado de proteccion
- Mapa interactivo con ubicacion de todos los servidores
- Lista de servidores filtrable por pais, ciudad y nombre
- Indicadores de carga por servidor con codificacion de colores (verde < 50%, amarillo 50-80%, rojo > 80%)
- Velocidad de subida/bajada en tiempo real
- Acciones rapidas: Quick Connect, Secure Core, P2P, Random
- Estadisticas de uso: datos transferidos, tiempo de conexion, sesiones totales
- Configuraciones: protocolo preferido, Kill Switch, Auto-Connect, proteccion DNS

---

## 🛡️ 4. Seguridad y Privacidad

### 🔐 4.1 Cifrado

| Capa | Tecnologia | Detalle |
|------|-----------|---------|
| Transporte API | TLS 1.3 | Comunicacion cliente-servidor cifrada |
| Tunel VPN (WireGuard) | ChaCha20-Poly1305 | Cifrado simetrico de alto rendimiento |
| Tunel VPN (OpenVPN) | AES-256-GCM | Cifrado de grado militar |
| Tunel VPN (IKEv2) | AES-256-GCM | Con Perfect Forward Secrecy |
| Intercambio de claves | Curve25519 | Criptografia de curva eliptica |
| Contrasenas | Argon2id | Hashing resistente a ataques de fuerza bruta |

### 🔒 4.2 Funciones de Proteccion

- 🚫 **Kill Switch** - Bloquea todo el trafico de internet si la VPN se desconecta inesperadamente
- 🌐 **Proteccion contra fugas DNS** - Fuerza todas las consultas DNS a traves del tunel cifrado
- 🔗 **Prevencion de fugas IPv6** - Enruta IPv6 por el tunel o lo bloquea
- 🛡️ **Proteccion WebRTC** - Bloquea servidores STUN que podrian revelar la IP real
- ⚡ **Split Tunneling** - Permite elegir que aplicaciones pasan por la VPN
- 🔄 **Multi-Hop (Double VPN)** - Encadena el trafico a traves de multiples nodos
- 🔒 **Ofuscacion** - Disfraza el trafico VPN como HTTPS para evadir censura
- 📝 **Politica No-Logs** - Cero registros de conexion o actividad

### 🔑 4.3 Proteccion del API

- Rate limiting: 100 req/min (anonimo), 1000 req/min (autenticado), 5000 req/min (premium)
- Headers de seguridad: HSTS, CSP, X-Frame-Options, X-Content-Type-Options
- CORS configurado por origenes permitidos

---

## 💰 5. Modelo Economico: Operadores de Nodos

### 💸 5.1 Como Funciona

Cualquier persona o empresa puede operar un nodo VPN en la red Aureo y **ganar criptomonedas** por cada gigabyte de trafico que procesan.

```
                    +------------------+
                    |  OPERADOR CREA   |
                    |  CUENTA + WALLET |
                    +--------+---------+
                             |
                    +--------v---------+
                    |  VERIFICACION    |
                    |  (Admin review)  |
                    +--------+---------+
                             |
                    +--------v---------+
                    |  DESPLIEGA NODO  |
                    |  VPN EN SERVIDOR |
                    +--------+---------+
                             |
                    +--------v---------+
                    |  USUARIOS SE     |
                    |  CONECTAN        |
                    +--------+---------+
                             |
                    +--------v---------+
                    |  GANANCIAS SE    |
                    |  ACUMULAN        |
                    +--------+---------+
                             |
                    +--------v---------+
                    |  PAGO EN CRYPTO  |
                    |  (ETH/BTC/LTC)   |
                    +------------------+
```

### 🏆 5.2 Sistema de Niveles de Recompensa

| Nivel | Tarifa por GB | Uptime Minimo | Reputacion Minima | Multiplicador |
|-------|--------------|---------------|-------------------|---------------|
| 🥉 Bronze | $0.010 | 50% | 0 | 1.0x |
| 🥈 Silver | $0.015 | 80% | 60 | 1.2x |
| 🥇 Gold | $0.020 | 90% | 75 | 1.5x |
| 💎 Platinum | $0.030 | 95% | 90 | 2.0x |

### 📊 5.3 Calculo de Ganancias

```
Ganancias = Bandwidth(GB) x Tarifa x Multiplicador de Calidad x Bonus de Duracion

Multiplicador de Calidad = 0.5 + (puntuacion_calidad / 100)
Bonus de Duracion = 1.1x (sesion > 60min) o 1.2x (sesion > 180min)
```

**Ejemplo:** Un nodo Platinum que sirve 100 GB con calidad del 90% en sesiones largas:
```
100 GB x $0.03 x (0.5 + 0.9) x 1.2 = $5.04 por esas sesiones
```

### ⭐ 5.4 Sistema de Reputacion

La reputacion de un operador se calcula automaticamente:

| Factor | Peso Maximo | Calculo |
|--------|-------------|---------|
| Base | 50 pts | Fijo para todos |
| Uptime | 30 pts | 30 x (uptime_promedio / 100) |
| Valoraciones de usuarios | 20 pts | 20 x (rating_promedio / 5.0) |
| Ancho de banda servido | 10 pts | Basado en GB totales |
| Staking (deposito de garantia) | 10 pts | Basado en monto bloqueado |
| **Total maximo** | **100 pts** | |

### 💳 5.5 Pagos

- **Pago minimo:** $10 USD
- **Criptomonedas:** Ethereum (ETH), Bitcoin (BTC), Litecoin (LTC)
- **Frecuencia:** Bajo demanda (el operador solicita cuando quiere)
- **Conversion:** Tipo de cambio en tiempo real via Price Oracle
- **Seguridad:** Confirmaciones en blockchain (ETH: 1, BTC: 3, LTC: 6)

### 📊 5.6 Dashboard del Operador

Los operadores tienen acceso a un panel completo con:
- Ingresos totales y pendientes
- Historial de ganancias por sesion
- Historial de pagos con hash de transaccion
- Estadisticas de nodos (uptime, conexiones, ancho de banda)
- Puntuacion de reputacion y nivel actual

---

## 🔐 6. Protocolos VPN Soportados

### 🔒 WireGuard (Recomendado)
- Protocolo moderno, rapido y ligero
- Cifrado ChaCha20-Poly1305
- Solo ~4,000 lineas de codigo (mas facil de auditar)
- Menor latencia y mayor rendimiento
- Intercambio de claves Curve25519

### 🔒 OpenVPN
- Protocolo establecido y ampliamente probado
- Cifrado AES-256-GCM con SHA256 HMAC
- Certificados RSA 2048-bit
- Compatible con la mayoria de firewalls

### 🔒 IKEv2/IPsec
- Ideal para dispositivos moviles (MOBIKE para roaming)
- Cifrado AES-256-GCM
- Perfect Forward Secrecy (PFS)
- Reconexion automatica al cambiar de red

---

## 🔧 7. Stack Tecnologico

| Componente | Tecnologia | Motivo |
|------------|-----------|--------|
| Backend | Go 1.24+ | Alto rendimiento, concurrencia nativa |
| Framework HTTP | Fiber v2 | Rendimiento superior a Express/Gin |
| Base de datos | SQLite | Embebida, cero configuracion, portable |
| ORM | GORM v2 | Migraciones automaticas, tipo-seguro |
| Autenticacion | JWT (HS256) | Stateless, escalable |
| Protocolo VPN | WireGuard | Moderno, rapido, seguro |
| Red P2P | libp2p | Descubrimiento descentralizado |
| Blockchain | go-ethereum | Pagos en criptomonedas |
| Metricas | Prometheus | Monitoreo en tiempo real |
| Cliente Desktop | Wails 2.11 | Nativo, ligero (vs Electron) |
| Frontend | Vanilla JS + Leaflet | Rapido, sin dependencias pesadas |
| CI/CD | GitHub Actions | Integracion continua automatizada |
| Contenedores | Docker | Despliegue estandarizado |

---

## 🏆 8. Ventajas Competitivas

### ⚔️ vs NordVPN / ExpressVPN / Surfshark

| Aspecto | VPNs Centralizadas | Aureo VPN |
|---------|-------------------|-----------|
| Infraestructura | Servidores propios, centralizados | Red descentralizada de nodos independientes |
| Codigo | Cerrado, propietario | Abierto, auditable |
| Confianza | "Confia en nosotros" | Verifica tu mismo |
| Censura | Un objetivo facil de bloquear | Miles de nodos independientes, dificil de bloquear |
| Modelo de negocio | Solo suscripcion | Suscripcion + oportunidad de ingreso para operadores |
| Punto de fallo | Unico proveedor | Sin punto unico de fallo |

### ⚔️ vs ProtonVPN

| Aspecto | ProtonVPN | Aureo VPN |
|---------|----------|-----------|
| Rendimiento | Bueno | Superior (WireGuard optimizado) |
| Ofuscacion | Basica | Avanzada (trafico disfrazado como HTTPS) |
| Economia | Solo consumidor | Consumidor + operador de nodos |
| Descentralizacion | Centralizado (Proton AG) | Verdaderamente descentralizado (P2P) |

---

## 💼 9. Casos de Uso

### 👤 Para Usuarios Individuales
- Proteger la privacidad en redes WiFi publicas
- Acceder a contenido restringido geograficamente
- Evitar censura en paises restrictivos (con ofuscacion)
- Proteger comunicaciones sensibles
- Navegar sin seguimiento de ISPs ni terceros

### 🏢 Para Empresas
- VPN corporativa autogestionada con nodos propios
- Acceso seguro remoto a recursos internos
- Cumplimiento de regulaciones de proteccion de datos
- Control total sobre la infraestructura de red
- Auditabilidad completa (codigo abierto)

### 📡 Para Operadores de Nodos
- Generar ingresos pasivos con servidores existentes
- Monetizar ancho de banda no utilizado
- Contribuir a una internet mas privada y libre
- Participar en una economia descentralizada

---

## 💳 10. Planes de Suscripcion

| Caracteristica | Free | Basic | Premium |
|---------------|------|-------|---------|
| Servidores disponibles | Limitados | Todos | Todos + dedicados |
| Ancho de banda | Limitado | Ilimitado | Ilimitado |
| Protocolos | WireGuard | WireGuard + OpenVPN | Todos (+ IKEv2) |
| Multi-Hop | --- | --- | ✅ |
| Dispositivos simultaneos | 1 | 3 | Ilimitados |
| Kill Switch | ✅ | ✅ | ✅ |
| Proteccion DNS | ✅ | ✅ | ✅ |
| Soporte | Comunidad | Email | Prioritario |
| Rate limit API | 100 req/min | 1,000 req/min | 5,000 req/min |

---

## 🚀 11. Despliegue y Operacion

### 📡 Para Operadores de Nodos

```bash
# 1. Registrarse como operador via API o CLI
aureo-cli operator register --wallet 0x... --wallet-type ethereum

# 2. Crear un nodo
aureo-cli node create \
  --name "us-east-1" \
  --hostname "vpn.midominio.com" \
  --country "US" --city "New York" \
  --max-connections 500

# 3. Ejecutar el nodo
./vpn-node --config config.yaml

# 4. Monitorear ganancias
aureo-cli operator stats
aureo-cli operator earnings
```

### 🔧 Para Administradores del Sistema
- Docker y Docker Compose disponibles
- Pipeline CI/CD con GitHub Actions
- Metricas Prometheus para monitoreo
- SQLite embebido (sin base de datos externa)
- Health checks: `/health` y `/ready`
- Metricas: `/metrics` (formato Prometheus)

---

## 🔒 12. Seguridad Demostrada

### 🔐 Cifrado de Extremo a Extremo
Cada conexion utiliza cifrado ChaCha20-Poly1305 (WireGuard) o AES-256-GCM (OpenVPN/IKEv2), los mismos estandares utilizados por agencias gubernamentales y entidades financieras.

### 🛡️ Sin Registros de Actividad
La arquitectura esta disenada para no almacenar registros de actividad de navegacion. Los unicos datos almacenados son los necesarios para el calculo de ganancias de operadores (bytes transferidos, no destinos visitados).

### 📖 Codigo Abierto
Todo el codigo es publico y auditable. Cualquier investigador de seguridad puede verificar las practicas de privacidad y reportar vulnerabilidades.

### 🌐 Descentralizacion Real
No existe un unico punto de control. La red P2P basada en libp2p con Kademlia DHT garantiza que ningun actor individual puede comprometer toda la red.

---

## 🗺️ 13. Roadmap Futuro

- 🚧 Aplicaciones moviles (iOS / Android)
- 🔮 Token nativo de Aureo en Solana para pagos y gobernanza
- 🔮 Sistema de gobernanza descentralizada (DAO)
- 🔮 Nodos de salida en mas de 100 paises
- 🔮 Integracion con wallets populares (MetaMask, Phantom)
- 🔮 Marketplace de servicios de privacidad adicionales
- 🚀 VPN para IoT y routers
- 🚀 API empresarial con SLAs garantizados

---

## 📬 Contacto

Para mas informacion sobre Aureo VPN, integraciones empresariales o como convertirse en operador de nodos, contactar al equipo de Aureo.

---

*Aureo VPN — Privacidad real. Descentralizada. Recompensada.*
