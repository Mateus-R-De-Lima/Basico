package main // Declara o pacote principal do programa Go, onde a execução começa

import (
	"encoding/json"
	"fmt"      // Pacote para formatação e impressão de saída (usado para escrever na resposta HTTP)
	"net/http" // Pacote para funcionalidades HTTP, incluindo servidor e manipulação de requisições/respostas
	"strconv"
	"time" // Pacote para manipulação de datas e horas (usado para obter a hora atual)

	"github.com/go-chi/chi/v5"            // Framework web Chi para roteamento HTTP, permite definir rotas e middlewares
	"github.com/go-chi/chi/v5/middleware" // Middlewares do Chi, como logger, recoverer, etc.
)

type Response struct {
	Error any `json:"error,omitempty"` // Campo Error do tipo any para armazenar informações de erro, com tag JSON para omitir se estiver vazio
	Data  any `json:"data,omitempty"`  // Campo Data do tipo any para armazenar dados de resposta, com tag JSON para omitir se estiver vazio
}

func sendJSON(w http.ResponseWriter, status int, res Response) { // Função auxiliar para enviar respostas JSON, recebe o ResponseWriter, status HTTP e os dados a serem enviados
	w.Header().Set("Content-Type", "application/json") // Define o header Content-Type como JSON
	data, err := json.Marshal(res)
	if err != nil {
		http.Error(w, "Error creating JSON response", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(status)

	if _, err := w.Write(data); err != nil {
		fmt.Println("Error writing response:", err)
		return
	}

}

type User struct { // Define uma estrutura (struct) chamada User para representar um usuário
	ID       uint64 `json:"id,string"` // Campo ID do tipo uint64 para armazenar o identificador do usuário, com tag JSON para serialização como string
	Name     string `json:"name"`      // Campo Name do tipo string para armazenar o nome do usuário, com tag JSON para serialização
	Role     string `json:"role"`      // Campo Role do tipo string para armazenar o papel ou função do usuário, com tag JSON para serialização
	Password string `json:"-"`         // Campo Password do tipo string para armazenar a senha do usuário, com tag JSON para omitir na serialização
}

func main() { // Função principal onde o programa Go inicia sua execução

	r := chi.NewMux() // Cria um novo multiplexador (router) usando o framework Chi para gerenciar rotas

	r.Use(middleware.Recoverer) // Adiciona middleware que recupera o servidor de panics, evitando que o programa pare abruptamente
	r.Use(middleware.RequestID) // Adiciona middleware que gera um ID único para cada requisição HTTP, útil para rastreamento
	r.Use(middleware.Logger)    // Adiciona middleware que registra logs das requisições HTTP no console

	r.Get("/horario", func(w http.ResponseWriter, r *http.Request) { // Define uma rota GET para o endpoint "/horario"
		now := time.Now()    // Obtém a data e hora atuais do sistema
		sendJSON(w, http.StatusOK, Response{Data: map[string]any{"horario": now}}) // Envia a hora atual em JSON
	})

	db := map[int64]User{
		1: {ID: 1, Name: "Alice", Role: "admin", Password: "secret"},
		2: {ID: 2, Name: "Bob", Role: "user", Password: "password"},
	}

	r.Group(func(r chi.Router) { // Cria um grupo de rotas para organizar endpoints relacionados
		r.Get("/users/{userId:[0-9]+}", handleGetUser(db)) // Define rota GET para "/users/{userId}" onde userId deve ser um número, usando a função handleGetUser para lidar com a requisição
		r.Post("/users", handlePostUser(db))               // Define rota POST para "/users" usando a função handlePostUser para lidar com a criação de um novo usuário (ainda sem implementação)

	})

	r.Route("/api", func(r chi.Router) { // Define um grupo de rotas aninhadas sob o prefixo "/api"
		r.Route("/v1", func(r chi.Router) { // Cria um subgrupo de rotas para a versão v1 da API
			r.Get("/users", func(w http.ResponseWriter, r *http.Request) {}) // Define rota GET para "/api/v1/users" (handler vazio, sem implementação ainda)
		})

		r.Route("/v2", func(r chi.Router) { // Cria um subgrupo de rotas para a versão v2 da API (atualmente vazio)
		})

		r.With(middleware.RealIP).Get("/users", func(w http.ResponseWriter, r *http.Request) {}) // Define rota GET para "/api/users" aplicando middleware RealIP (obtém IP real do cliente)

		r.With(middleware.RealIP).Get("/users/{userId:[0-9]+}", func(w http.ResponseWriter, r *http.Request) { // Define rota GET para "/api/users/{userId}" onde userId deve ser um número, aplicando middleware RealIP
			id := chi.URLParam(r, "userId")             // Extrai o parâmetro "userId" da URL usando a função URLParam do Chi
			fmt.Fprintln(w, "Olá, usuário com ID :"+id) // Escreve uma resposta personalizada usando o ID do usuário extraído da URL
		})

		r.Group(func(r chi.Router) { // Cria um grupo de rotas que compartilha middlewares
			r.Use(middleware.BasicAuth("", map[string]string{ // Adiciona middleware de autenticação básica HTTP
				"admin": "admin", // Define credenciais: usuário "admin" com senha "admin"
			}))

			r.Get("/healthcheck", func(w http.ResponseWriter, r *http.Request) { // Define rota GET para "/api/healthcheck" dentro do grupo (requer auth)
				fmt.Fprintln(w, "ping") // Responde com "ping" para indicar que o serviço está saudável
			})
		})
	})

	if err := http.ListenAndServe(":8080", r); err != nil { // Inicia o servidor HTTP na porta 8080 usando o router Chi
		panic(err) // Se houver erro ao iniciar o servidor, causa um pânico (encerra o programa com erro)
	}
}

func handleGetUser(db map[int64]User) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "userId")       // Extrai o parâmetro "userId" da URL usando a função URLParam do Chi
		id, _ := strconv.ParseInt(idStr, 10, 64) // Converte o ID de string para int64
		user, ok := db[id]

		if ok {
			sendJSON(w, http.StatusOK, Response{Data: user}) // Usa sendJSON para enviar o usuário em formato JSON
		} else {
			sendJSON(w, http.StatusNotFound, Response{Error: "Usuário não encontrado"}) // Usa sendJSON para enviar erro
		}
	}
}

func handlePostUser(db map[int64]User) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1000) // Limita o tamanho do corpo da requisição para evitar abusos

		var user User
		decoder := json.NewDecoder(r.Body) // Cria um decoder JSON para ler o corpo da requisição
		decoder.DisallowUnknownFields()    // Rejeita campos extras que não existem na struct User

		if err := decoder.Decode(&user); err != nil { // Deserializa o JSON diretamente no struct User
			sendJSON(w, http.StatusUnprocessableEntity, Response{Error: "Body inválido: JSON inválido ou campos extras não permitidos"})
			return
		}

		if errs := validateUser(user); len(errs) > 0 { // Valida os campos obrigatórios do usuário
			sendJSON(w, http.StatusUnprocessableEntity, Response{Error: map[string]any{"erros": errs}})
			return
		}

		db[int64(user.ID)] = user                    // Adiciona o novo usuário ao "banco de dados" (mapa em memória)
		sendJSON(w, http.StatusCreated, Response{Data: map[string]string{"message": "Usuário criado com sucesso"}}) // Usa sendJSON para resposta de criação
	}
}

func validateUser(user User) []map[string]string {
	errs := []map[string]string{}

	if user.ID == 0 {
		errs = append(errs, map[string]string{"Id": "É obrigatório informar o Id"})
	}

	if user.Name == "" {
		errs = append(errs, map[string]string{"Nome": "É obrigatório informar o nome"})
	}

	if user.Role == "" {
		errs = append(errs, map[string]string{"Role": "É obrigatório informar o role"})
	}

	return errs
}
