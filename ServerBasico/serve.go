package serverbasico

// Aqui estamos dizendo quais ferramentas o programa vai usar.
// - tls: pra internet segura
// - errors: pra entender erros
// - log: pra escrever mensagens na tela
// - http: pra criar um servidor web
// - time: pra contar tempo
import (
	"crypto/tls"
	"errors"
	"log"
	"net/http"
	"time"
)

func exemplo() {

	// Esta é a porta de entrada do programa. Tudo começa aqui.

	mux := http.NewServeMux()
	// Criamos um "coletor de rotas", um lugar onde guardamos
	// quais endereços o servidor deve responder.

	mux.HandleFunc("/api/users/{userId}", func(w http.ResponseWriter, r *http.Request) {
		// Aqui dizemos: "Quando alguém acessar /api/users/{id},
		// responda com 'Hola Mundo'."
		// w é o lugar onde escrevemos a resposta.
		// r é o pedido que chegou do visitante.
		id := r.PathValue("userId")
		w.Write([]byte("Olá, usuário com ID :" + id))
	})
	// Aqui dizemos: "Quando alguém acessar /healthcheck,
	// responda com 'Hola Mundo'."
	// w é o lugar onde escrevemos a resposta.
	// r é o pedido que chegou do visitante.

	servidor := &http.Server{
		Addr:                         "localhost:8080",
		Handler:                      mux,
		DisableGeneralOptionsHandler: false,
		TLSConfig:                    &tls.Config{},
		ReadTimeout:                  30 * time.Second,
		WriteTimeout:                 10 * time.Second,
		IdleTimeout:                  1 * time.Minute,
		MaxHeaderBytes:               0,
	}
	// Aqui criamos o servidor.
	// - Addr: onde ele vai escutar (localhost na porta 8080)
	// - Handler: o lugar que decide o que responder
	// - TLSConfig: configuração de segurança (aqui está vazia)
	// - ReadTimeout: tempo máximo para ler um pedido
	// - WriteTimeout: tempo máximo para enviar a resposta
	// - IdleTimeout: tempo de espera quando nada acontece
	// - MaxHeaderBytes: tamanho máximo dos cabeçalhos

	log.Fatal(servidor.ListenAndServe())
	// Este comando inicia o servidor de verdade.
	// Ele fica esperando alguém visitar o site.
	// Se der algum erro, ele escreve o erro e fecha o programa.

	if err := servidor.ListenAndServe(); err != nil {
		// Este bloco tenta iniciar o servidor de novo e
		// trata o erro se houver.
		// Mas o código acima já terminou o programa se houve erro,
		// então este bloco nunca será alcançado.

		if !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
		// Se o erro não for o de servidor fechado,
		// então o programa para com um erro grande.
	}

}
