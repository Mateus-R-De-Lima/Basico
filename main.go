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
		fmt.Fprintln(w, now) // Escreve a hora atual na resposta HTTP, enviando para o cliente
	})

	db := map[int64]User{
		1: {ID: 1, Name: "Alice", Role: "admin", Password: "secret"},
		2: {ID: 2, Name: "Bob", Role: "user", Password: "password"},
	}

	r.Group(func(r chi.Router) { // Cria um grupo de rotas para organizar endpoints relacionados
		r.Use(jsonMiddleware)                              // Adiciona middleware personalizado para definir o header Content-Type como application/json para todas as rotas dentro deste grupo
		r.Get("/users/{userId:[0-9]+}", handleGetUser(db)) // Define rota GET para "/users/{userId}" onde userId deve ser um número, usando a função handleGetUser para lidar com a requisição
		r.Post("/users", handlePostUser)                   // Define rota POST para "/users" usando a função handlePostUser para lidar com a criação de um novo usuário (ainda sem implementação)

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

func jsonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func handleGetUser(db map[int64]User) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "userId")       // Extrai o parâmetro "userId" da URL usando a função URLParam do Chi
		id, _ := strconv.ParseInt(idStr, 10, 64) // Converte o ID de string para int64
		user, ok := db[id]

		if ok {
			data, err := json.Marshal(user) // Serializa o usuário para JSON
			if err != nil {
				http.Error(w, "Erro ao serializar usuário", http.StatusInternalServerError) // Retorna erro se a serialização falhar
				return
			}
			w.Write(data) // Envia os dados do usuário serializados como resposta HTTP
		} else {
			http.Error(w, "Usuário não encontrado", http.StatusNotFound) // Retorna erro se o usuário não for encontrado
		}
	}
}

func handlePostUser(w http.ResponseWriter, r *http.Request) {}
