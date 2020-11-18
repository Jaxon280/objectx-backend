package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/boltdb/bolt"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"gocv.io/x/gocv"
)

const (
	DBPath   = "main.db"
	CDNHost  = "https://objectx.ams3.cdn.digitaloceanspaces.com"
	Endpoint = "ams3.digitaloceanspaces.com"
)

type Server struct {
	db       *bolt.DB
	s3Client *s3.S3
}

type ImageJSON struct {
	ID         string  `json:"id"`
	Sentiment  float64 `json:"sentiment,omitempty"`
	Brightness float64 `json:"brightness,omitempty"`
	Tone       float64 `json:"tone,omitempty"`
	Created    int64   `json:"created,omitempty"`
	ImageURL   string  `json:"image_url,omitempty"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

func analyzeImage(filepath string) (float64, float64) {
	mat := gocv.IMRead(filepath, gocv.IMReadColor)
	defer mat.Close()

	gocv.CvtColor(mat, &mat, gocv.ColorBGRToHSV)
	hsv := gocv.Split(mat)

	return hsv[0].Mean().Val1 * 2.0, hsv[2].Mean().Val1 // color tone(multiple by 2 to normalize), brightness
}

func decode(buffer []byte) (*ImageJSON, error) {
	var image ImageJSON
	err := json.Unmarshal(buffer, &image)
	if err != nil {
		return nil, err
	}

	return &image, nil
}

func (s *Server) writeError(w http.ResponseWriter, message string, status int) {
	resp := ErrorResponse{
		Message: message,
	}
	v, _ := json.Marshal(resp)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if size, err := w.Write(v); err != nil || size == 0 {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) createBoltBucket(name string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(name))
		return err
	})
}

func (s *Server) close() error {
	if err := s.db.Close(); err != nil {
		return err
	}
	return nil
}

func (s *Server) readImages(max int) (map[string]ImageJSON, error) {
	tx, err := s.db.Begin(false)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer tx.Rollback()

	buffer := make(map[string]ImageJSON, max)
	b := tx.Bucket([]byte("image"))
	c := b.Cursor()
	counter := 0

	for k, v := c.First(); k != nil && counter < max; k, v = c.Next() {
		image, err := decode(v)
		if err != nil {
			return nil, err
		}

		buffer[string(k)] = *image
		counter++
	}
	return buffer, nil
}

func (s *Server) readImage(key []byte) ([]byte, error) {
	tx, err := s.db.Begin(false)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer tx.Rollback()

	b := tx.Bucket([]byte("image"))
	v := b.Get(key)
	if v == nil {
		return nil, fmt.Errorf("image (id: %s) is not found.", key)
	}

	if err != nil {
		return nil, err
	}

	return v, nil
}

func (s *Server) createImage(image *ImageJSON) ([]byte, error) {
	tx, err := s.db.Begin(true)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer tx.Rollback()

	b := tx.Bucket([]byte("image"))

	bytes, err := json.Marshal(image)
	if err != nil {
		return nil, err
	}

	if err := b.Put([]byte(image.ID), bytes); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return bytes, nil
}

func (s *Server) destroyImage(key []byte) ([]byte, error) {
	tx, err := s.db.Begin(true)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer tx.Rollback()

	b := tx.Bucket([]byte("image"))
	if err := b.Delete(key); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	v, err := json.Marshal(&ImageJSON{
		ID:         string(key),
		Sentiment:  0,
		Brightness: 0,
		Tone:       0,
		Created:    0,
		ImageURL:   "",
	})
	if err != nil {
		return nil, err
	}

	return v, nil
}

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if r.Method != http.MethodGet {
		s.writeError(w, "Only GET method is allowed.", http.StatusMethodNotAllowed)
		return
	}

	buf, err := s.readImages(40)
	if err != nil {
		s.writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	bytes, err := json.MarshalIndent(buf, "", "	")
	if err != nil {
		s.writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(bytes); err != nil {
		s.writeError(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) uploadHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if r.Method != http.MethodPost {
		s.writeError(w, "Only POST method is allowed.", http.StatusMethodNotAllowed)
		return
	}

	image := ImageJSON{}
	image.ID = uuid.New().String()

	file, reader, err := r.FormFile("image")
	if err != nil {
		s.writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	tmpfile, err := ioutil.TempFile("", "temp"+filepath.Ext(reader.Filename))
	if err != nil {
		s.writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer os.Remove(tmpfile.Name())

	if _, err := io.Copy(tmpfile, file); err != nil {
		s.writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ext := filepath.Ext(reader.Filename)
	image_id := uuid.New().String() + ext

	uploadfile, err := reader.Open()
	defer uploadfile.Close()

	object := s3.PutObjectInput{
		Bucket:        aws.String("objectx"),
		Key:           aws.String(image_id),
		Body:          uploadfile,
		ContentLength: aws.Int64(reader.Size),
		ACL:           aws.String("public-read"),
	}

	if _, err := s.s3Client.PutObject(&object); err != nil {
		fmt.Println(err.Error())
		s.writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	image.ImageURL = CDNHost + "/" + image_id

	image.Tone, image.Brightness = analyzeImage(tmpfile.Name())
	if err := tmpfile.Close(); err != nil {
		s.writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sentiment, err := strconv.ParseFloat(r.FormValue("sentiment"), 64)
	if err != nil {
		image.Sentiment = 0
	} else {
		image.Sentiment = sentiment
	}

	image.Created = time.Now().Unix()

	js, err := s.createImage(&image)
	if err != nil {
		s.writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if _, err := w.Write(js); err != nil {
		s.writeError(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) pickHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if r.Method != http.MethodGet {
		s.writeError(w, "Only GET method is allowed.", http.StatusMethodNotAllowed)
		return
	}

	// - IDをパスから取得
	// - boltDBからIDを元に読み込み
	// - レスポンスに書き込み
	targetID := ps.ByName("id")

	js, err := s.readImage([]byte(targetID))
	if err != nil {
		s.writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(js); err != nil {
		s.writeError(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) deleteHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if r.Method != http.MethodDelete {
		s.writeError(w, "Only DELETE method is allowed.", http.StatusMethodNotAllowed)
		return
	}

	targetID := ps.ByName("id")
	js, err := s.destroyImage([]byte(targetID))
	if err != nil {
		s.writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(js); err != nil {
		s.writeError(w, err.Error(), http.StatusInternalServerError)
	}
}

func newServer(dbPath string) (*Server, error) {
	db, err := bolt.Open(dbPath, 0666, nil)
	if err != nil {
		log.Fatal("open db")
		return nil, err
	}

	key := os.Getenv("SPACES_KEY")       // set the key
	secret := os.Getenv("SPACES_SECRET") // set the secret

	s3Config := &aws.Config{
		Credentials: credentials.NewStaticCredentials(key, secret, ""),
		Endpoint:    aws.String("https://" + Endpoint),
		Region:      aws.String("us-east-1"),
	}
	newSession := session.New(s3Config)
	cli := s3.New(newSession)

	return &Server{
		db:       db,
		s3Client: cli,
	}, nil
}

func main() {
	err := os.Chdir(filepath.Join("/root", "go", "src", "objectx-backend")) // if in local ENV, os.Chdir(filepath.Join("/home", "environment", "go", "src", "objectx-backend"))
	if err != nil {
		log.Fatal(err.Error())
		os.Exit(1)
	}

	server, err := newServer(DBPath)
	if err != nil {
		log.Fatal(err.Error())
		os.Exit(1)
	}
	defer server.close()

	if err := server.createBoltBucket("image"); err != nil && err != bolt.ErrBucketExists {
		log.Fatal(err.Error())
		os.Exit(1)
	}

	router := httprouter.New()
	router.GET("/images", server.indexHandler)
	router.POST("/upload", server.uploadHandler)
	router.GET("/images/:id", server.pickHandler)
	router.DELETE("/images/:id", server.deleteHandler)

	fmt.Println("Start Server...")

	log.Fatal(http.ListenAndServe(":80", router))
}
