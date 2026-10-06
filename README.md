# mundoappoinment in go

Fullstack application that helps manage medical appointments and therapy sessions for all healthcare personnel, while also providing key insights and metrics to help optimize the business side of operations. The idea is to give flexibility to healtcare personnel to managed their own agenda or have the visibility of multiple agendas in case of clinics.

## Preview


## Architecture

Diseño del siguiente módulo: [usuarios, membresías y profesionales](docs/usuarios-membresias-profesionales.md).
Diseño mínimo de agenda: [citas, disponibilidad y capacidad](docs/cita-disponibilidad.md).

### Backend
The backend structure is organized by domain, and each domain is divided into layers consisting of:

- **Routes**: Mapping of endpoint and handler
- **Handlers**: Entre point of any requests and responsable to send a response to client
- **Store**: Responsable to send request to DB to fetch data

### API

**Patient**
Patient admission date


## Gestión del proyecto

El [plan único en Notion](https://app.notion.com/p/46deaaf04e474ebab434087ae36be978)
contiene la secuencia, las dependencias, los estados y los objetivos para planificar
la semana. Usa **Secuencia** para consultar el orden, **Por estado** para revisar
avances y **Para planificar** para seleccionar el siguiente trabajo.

La [página del proyecto](https://app.notion.com/p/3e45a22648758095ab3cf338b32a1aeb)
explica cómo mantener el tablero y preparar objetivos de calendario. No existe
sincronización automática: antes de planificar, revisar avances y dependencias;
acordar disponibilidad y fechas antes de crear eventos.

Los Markdown conservan las reglas y el esquema técnico, no una segunda lista de
tareas. El antiguo TODO queda consolidado en el tablero: CRUD y alcance por
clínica en BASE-01/BASE-02, CI en BASE-03, seguridad en SEG-03, pruebas adicionales
en QA-01 y frontend en FUT-01. La revisión de Gin figuraba como terminada en el
TODO histórico; no representa una validación nueva. CI no implica despliegue.

Referencia para consultar el tablero mediante la integración de Notion:
data source `233fa4b0-b60e-4e68-a467-a837bb33644a`.
