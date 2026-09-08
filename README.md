# mundoappoinment in go

Fullstack application that helps manage medical appointments and therapy sessions for all healthcare personnel, while also providing key insights and metrics to help optimize the business side of operations. The idea is to give flexibility to healtcare personnel to managed their own agenda or have the visibility of multiple agendas in case of clinics.

## Preview


## Architecture

### Backend
The backend structure is organized by domain, and each domain is divided into layers consisting of:

- **Routes**: Mapping of endpoint and handler
- **Handlers**: Entre point of any requests and responsable to send a response to client
- **Store**: Responsable to send request to DB to fetch data

### API

**Patient**
Patient admission date


## TODO
- [ ] Init Frontend side
- [ ] Add more integration test to patients domain
- [ ] Verify from gin docs if any change is needed to imporve the code
- [ ] Add CI/CD
- [ ] Investigate cloud provider