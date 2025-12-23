# Escenarios de Datos de Prueba - Objetivos con Valor Inicial

## Resumen

Este documento describe los escenarios de prueba creados por el script `./zarf/seed-data.sh` para validar el funcionamiento de objetivos con valor inicial (begin value).

## Credenciales de Prueba

```
Email:    admin@rafiki.lat
Password: admin123
```

## Como Ejecutar

```bash
# Asegurate de tener Docker corriendo
make up

# Ejecutar el seed de datos
./zarf/seed-data.sh
```

---

## Escenarios de Objetivos Tipo Resultado

### 1. Objetivo de Incremento - 50% completado

| Campo | Valor |
|-------|-------|
| **Titulo** | Leer 100 libros este year |
| **Valor inicial** | 0 |
| **Meta** | 100 |
| **Actual** | 50 |
| **Progreso esperado** | 50% |
| **Texto esperado** | "50 de 100" |

**Que verificar:**
- La barra de progreso muestra 50%
- El texto dice "50 de 100 (50% completado)"
- El boton + aumenta el valor actual
- El boton - disminuye el valor actual

---

### 2. Objetivo de Reduccion - 0% completado (recien comenzado)

| Campo | Valor |
|-------|-------|
| **Titulo** | Reducir peso (50 a 30) |
| **Valor inicial** | 50 |
| **Meta** | 30 |
| **Actual** | 50 |
| **Progreso esperado** | 0% |
| **Texto esperado** | "0 reducido de 20" |

**Que verificar:**
- La barra de progreso muestra 0%
- El texto dice "0 reducido de 20 (0% completado)"
- Se muestra "Inicio: 50 ↓" indicando que es objetivo de reduccion
- El boton - reduce el valor (avanza hacia la meta)
- El boton + aumenta el valor (retrocede)

---

### 3. Objetivo de Reduccion - 50% completado

| Campo | Valor |
|-------|-------|
| **Titulo** | Reducir cafeina (80 a 70) |
| **Valor inicial** | 80 |
| **Meta** | 70 |
| **Actual** | 75 |
| **Progreso esperado** | 50% |
| **Texto esperado** | "5 reducido de 10" |

**Que verificar:**
- La barra de progreso muestra 50%
- El texto dice "5 reducido de 10 (50% completado)"
- Se muestra "Inicio: 80 ↓"

---

### 4. Objetivo de Reduccion - 100% completado

| Campo | Valor |
|-------|-------|
| **Titulo** | Reducir tiempo pantalla (100 a 80) |
| **Valor inicial** | 100 |
| **Meta** | 80 |
| **Actual** | 80 |
| **Progreso esperado** | 100% |
| **Texto esperado** | "20 reducido de 20" |

**Que verificar:**
- La barra de progreso muestra 100%
- El texto dice "20 reducido de 20 (100% completado)"
- El color de la barra es azul (progreso positivo)

---

### 5. Objetivo de Reduccion - Superado (overshot)

| Campo | Valor |
|-------|-------|
| **Titulo** | Reducir azucar (50 a 30) - superado |
| **Valor inicial** | 50 |
| **Meta** | 30 |
| **Actual** | 20 |
| **Progreso esperado** | 100% (meta superada) |
| **Texto esperado** | "20 reducido de 20" |

**Que verificar:**
- La barra de progreso muestra 100% (no mas)
- El valor actual (20) es menor que la meta (30)
- Esto indica que el usuario supero su objetivo

---

### 6. Objetivo de Incremento - 25% completado

| Campo | Valor |
|-------|-------|
| **Titulo** | Ahorrar 10000 pesos |
| **Valor inicial** | 0 |
| **Meta** | 10000 |
| **Actual** | 2500 |
| **Progreso esperado** | 25% |
| **Texto esperado** | "2500 de 10000" |

**Que verificar:**
- La barra de progreso muestra 25%
- El texto dice "2500 de 10000 (25% completado)"
- No se muestra "Inicio: X" porque valor inicial es 0

---

## Escenario de Objetivo Tipo Frecuencia

### 7. Meditacion diaria - 90 dias de historial

| Campo | Valor |
|-------|-------|
| **Titulo** | Meditar 10 minutos diarios |
| **Tipo** | Frecuencia (diaria) |
| **Meta cumplimiento** | 80% |
| **Registros** | 90 dias con ~80% completados |

**Que verificar:**
- El heatmap muestra actividad de los ultimos 90 dias
- Aproximadamente 80% de los dias estan marcados como completados
- El contador de racha muestra dias consecutivos
- El contador total muestra todos los dias completados

---

## Tabla Resumen de Calculos

| Escenario | Begin | Target | Current | Rango | Progreso | % |
|-----------|-------|--------|---------|-------|----------|---|
| Incremento normal | 0 | 100 | 50 | 100 | 50 | 50% |
| Reduccion inicio | 50 | 30 | 50 | 20 | 0 | 0% |
| Reduccion mitad | 80 | 70 | 75 | 10 | 5 | 50% |
| Reduccion completo | 100 | 80 | 80 | 20 | 20 | 100% |
| Reduccion superado | 50 | 30 | 20 | 20 | 20+ | 100% |
| Incremento 25% | 0 | 10000 | 2500 | 10000 | 2500 | 25% |

---

## Formula de Calculo

### Para objetivos de incremento (begin < target):
```
Rango total = target - begin
Progreso = current - begin
Porcentaje = (progreso / rango) * 100
Texto = "{progreso} de {rango}"
```

### Para objetivos de reduccion (begin > target):
```
Rango total = begin - target
Progreso = begin - current
Porcentaje = (progreso / rango) * 100
Texto = "{progreso} reducido de {rango}"
```

---

## Bugs Corregidos

### 1. Backend: Validacion de progreso para objetivos inversos
**Antes:** El sistema rechazaba actualizaciones donde `current > target` para TODOS los objetivos.
**Despues:** Solo se rechaza para objetivos de incremento. Para objetivos de reduccion, `current > target` es progreso valido (0-99%).

### 2. Backend: Inicializacion de currentMetric
**Antes:** Todos los objetivos iniciaban con `currentMetric = 0`.
**Despues:** Objetivos de reduccion inician con `currentMetric = beginMetric`.

### 3. Frontend: Texto de progreso
**Antes:** Mostraba "50 de 20" para reduccion (incorrecto).
**Despues:** Muestra "X reducido de Y" para reduccion.

---

## Tareas de Prueba Incluidas

El seed tambien crea tareas de ejemplo:

### Tareas de Inbox (sin objetivo):
- Revisar y responder emails
- Planificar la semana
- Llamar al dentista (completada)

### Tareas vinculadas a objetivos:
- Leer 30 paginas del libro actual (contribucion: 5, completada)
- Comprar 3 libros de la lista (contribucion: 3, pendiente)
- Caminata de 30 minutos (contribucion: 2, completada)
- Preparar comidas de la semana (contribucion: 5, pendiente)
