#### Structure

`auth.go`:
* Authmiddleware
* Role and user management endpoints
* Endpoint authentification system

`buyer.go`:
* Buyer related endpoints used by the Portal

`ops.go`:
* Endpoints setup to be used specifically by the ops tool

`*_test.go`:
* Unit tests for the corresponding service (auth, buyer, ops)