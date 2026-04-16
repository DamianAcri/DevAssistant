instalaré esto en backend/
go get github.com/go-chi/chi/v5 //esto es para lo del servidor http entiendo
go get github.com/joho/godotenv //esto es para cargar las variables de entorno desde un archivo .env

ahora entiendo que router deberia exponer NewRouter() http.handler 

la idea es esto: go run ./cmd/server
curl localhost:8080/health

y que responda: {"status":"ok"}

supongo que lo primero que tengo que hacer es el webhook reciever entiendo. para que? para que el servidor pueda recibir peticiones y responder a ellas. hare el /health primero porque es lo mas sencillo y luego el webhook receiver que es lo que realmente necesito para recibir las peticiones de github.

con := inferimos el tipo de dato autmaticamente

Ahora haré el handler de /health. los handlers son funciones encargados de manejar peticiones HTTP, como POST o GET, procesar la solicitud y envíar una res al cliente. uso net/http para esto, es una libreria estandar de go para manejar HTTP.

En el handler: json.NewEncoder(w).Encode(map[string]string{
		"status": "ok"
	})

esto es para enviar una respuesta en formato JSON al cliente, con un mapa que tiene una clave "status" y un valor "ok". w es el ResponseWriter que se usa para escribir la respuesta HTTP. json.NewEncoder(w) crea un nuevo encoder JSON que escribe directamente en el ResponseWriter, y Encode() codifica el mapa como JSON y lo envía al cliente. No ponemos un status code porque por defecto es 200 OK, que es lo que queremos para indicar que la solicitud fue exitosa.

La verdad es que el archivo de webhook es bastante descriptiivo por si mismo. 

ahora paso a router.go, que es donde configurare las rutas de mi servidor HTTP usando chi. chi es un router ligero para Go que facilita la creación de rutas y el manejo de solicitudes HTTP.

ahora hago el main.go, que es el punto de entrada de mi aplicación. aquí es donde inicializaré el router y arrancaré el servidor HTTP. cargo las variables de entorno con config, luego creo el router con api.NewRouter() y luego arranco el servidor con http.ListenAndServe, pasando el puerto y el router como argumentos. esto hará que el servidor escuche en el puerto especificado y maneje las solicitudes usando el router que configuré.

