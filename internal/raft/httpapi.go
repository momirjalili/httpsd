package raft

import (
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/momirjalili/httpsd/internal/api"
	bolt "go.etcd.io/bbolt"
	"go.etcd.io/etcd/raft/v3/raftpb"
)

// Handler for a http based key-value store backed by raft
type httpKVAPI struct {
	store       *KVStore
	confChangeC chan<- raftpb.ConfChange
}

func (h *httpKVAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	key := r.RequestURI
	defer r.Body.Close()
	switch r.Method {
	case http.MethodPut:
		v, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Failed to read on PUT (%v)\n", err)
			http.Error(w, "Failed on PUT", http.StatusBadRequest)
			return
		}

		h.store.Propose(key, string(v))

		// Optimistic-- no waiting for ack from raft. Value is not yet
		// committed so a subsequent GET on the key may return old value
		w.WriteHeader(http.StatusNoContent)
	case http.MethodGet:
		if v, ok := h.store.Lookup(key); ok {
			w.Write([]byte(v))
		} else {
			http.Error(w, "Failed to GET", http.StatusNotFound)
		}
	case http.MethodPost:
		url, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Failed to read on POST (%v)\n", err)
			http.Error(w, "Failed on POST", http.StatusBadRequest)
			return
		}

		nodeId, err := strconv.ParseUint(key[1:], 0, 64)
		if err != nil {
			log.Printf("Failed to convert ID for conf change (%v)\n", err)
			http.Error(w, "Failed on POST", http.StatusBadRequest)
			return
		}

		cc := raftpb.ConfChange{
			Type:    raftpb.ConfChangeAddNode,
			NodeID:  nodeId,
			Context: url,
		}
		h.confChangeC <- cc
		// As above, optimistic that raft will apply the conf change
		w.WriteHeader(http.StatusNoContent)
	case http.MethodDelete:
		nodeId, err := strconv.ParseUint(key[1:], 0, 64)
		if err != nil {
			log.Printf("Failed to convert ID for conf change (%v)\n", err)
			http.Error(w, "Failed on DELETE", http.StatusBadRequest)
			return
		}

		cc := raftpb.ConfChange{
			Type:   raftpb.ConfChangeRemoveNode,
			NodeID: nodeId,
		}
		h.confChangeC <- cc

		// As above, optimistic that raft will apply the conf change
		w.WriteHeader(http.StatusNoContent)
	default:
		w.Header().Set("Allow", http.MethodPut)
		w.Header().Add("Allow", http.MethodGet)
		w.Header().Add("Allow", http.MethodPost)
		w.Header().Add("Allow", http.MethodDelete)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// serveHttpKVAPI starts a key-value server with a GET/PUT API and listens.
func ServeHttpKVAPI(kv *KVStore, port int, confChangeC chan<- raftpb.ConfChange, errorC <-chan error) {
	srv := http.Server{
		Addr: ":" + strconv.Itoa(port),
		Handler: &httpKVAPI{
			store:       kv,
			confChangeC: confChangeC,
		},
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	// exit when raft goes down
	if err, ok := <-errorC; ok {
		log.Fatal(err)
	}
}

// serveHttpSDAPI starts a key-value server with a GET/PUT API and listens.
func ServeHttpSDAPI() {

	router := mux.NewRouter()
	router.StrictSlash(true)
	// db, err := storm.Open("my.db", storm.Codec(gob.Codec))
	db, err := bolt.Open("my.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	server := api.NewSDServer(db)

	router.HandleFunc("/api/v1/target/", server.GetAllTargetGroupsHandler).Methods("GET")
	router.HandleFunc("/api/v1/target/", server.CreateTargetGroupHandler).Methods("POST")
	router.HandleFunc("/api/v1/target/{id:[0-9]+}/", server.GetTargetGroupHandler).Methods("GET")
	router.HandleFunc("/api/v1/target/{id:[0-9]+}/", server.PutTargetHandler).Methods("PUT")
	router.HandleFunc("/api/v1/target/{id:[0-9]+}/label/{label_key}", server.PatchTargetGroupLabelHandler).Methods("PATCH")
	router.HandleFunc("/api/v1/target/{id:[0-9]+}/label/{label_key}", server.DeleteTargetGroupLabelHandler).Methods("DELETE")
	router.HandleFunc("/api/v1/target/{id:[0-9]+}/label/{server_key}", server.DeleteTargetGroupLabelHandler).Methods("DELETE")
	http.ListenAndServe(":8080", router)
}

// func getOrCreateDB() {}
