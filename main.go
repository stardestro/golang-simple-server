package main

import (
	"fmt"
	"net"
	"html/template"
	"net/http"
	"io/ioutil"
	"os"
	"io"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"log"
)

var tpl *template.Template

func main() {
	logfile, logerr := os.OpenFile("serve.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if logerr != nil {
		log.Fatal(logerr)
	}
	defer logfile.Close()
	log.SetOutput(logfile)
	tpl, _ = tpl.ParseGlob("/home/templates/*.html")
	// create our new var myDir at type http.Dir
	myDir := http.Dir("./public/")
	
	log.Printf("myDir type: %T", myDir)
	//encryptFile("sample.txt", []byte("Hello World"), "password1")
	//decryptFile("sample.txt", "password1")
	
	// func FileServer(root FileSystem) Handler
	//myHandler := http.FileServer(myDir)
	http.HandleFunc("/", index)
	// using absolute path
	// http.Handle("/", http.FileServer(http.Dir("/workspace/goworkspace/src/gowebdev/fileServer/public")))
	// using relative path
	// http.Handle("/", http.FileServer(http.Dir("./public")))
	// does not work, will look at ./public/public
	// http.Handle("/public", http.FileServer(http.Dir("./public")))
	// use http.StringPrefix to alter request before FileServer sees it
	// http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("./public"))))
	// http.Handle("/public/", http.FileServer(http.Dir(".")))
	// http.HandleFunc("/hello", helloHandler)
	http.HandleFunc("/download", downloadFile)
	http.HandleFunc("/upload", uploadFile)
	http.HandleFunc("/directory", directory)
	err := http.ListenAndServe(":80", nil)
	fmt.Println("%v", err)
	//http.HandleFunc("/cat", catfunction)
}

func index(w http.ResponseWriter, r *http.Request) {
	log.Println("r.method:", r.Method)
	IPAddress := r.Header.Get("X-Real-Ip")
        if IPAddress == "" {
                IPAddress = r.Header.Get("X-Forwarded-For")
        }
        if IPAddress == "" {
                IPAddress = r.RemoteAddr
        }
        log.Println(net.ParseIP(IPAddress))
	// if method is GET then load form, if not then upload successfull message
	if r.Method == "GET" {
		tpl.ExecuteTemplate(w, "index.html", nil)
		return
	}
}

func createHash(key string) string {
	hasher := md5.New()
	hasher.Write([]byte(key))
	return hex.EncodeToString(hasher.Sum(nil))
}

func encrypt(data []byte, passphrase string) []byte {
	block, _ := aes.NewCipher([]byte(createHash(passphrase)))
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext
}

func decrypt(data []byte, passphrase string) []byte {
	key := []byte(createHash(passphrase))
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}
	nonceSize := gcm.NonceSize()
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		panic(err.Error())
	}
	return plaintext
}

func encryptFile(filename string, data []byte, passphrase string) {
	f, _ := os.Create(filename)
	defer f.Close()
	f.Write(encrypt(data, passphrase))
}

func decryptFile(filename string, passphrase string) {
	data, _ := ioutil.ReadFile("/home/public/" + filename)
	/*
	
	*/
	f, _ := os.Create("/home/public/decrypted" + filename)
	defer f.Close()
	f.Write(decrypt(data, passphrase))
}

func downloadFile(w http.ResponseWriter, r *http.Request) {
	log.Println("r.method:", r.Method)
	// if method is GET then load form, if not then upload successfull message
	if r.Method == "GET" {
		files, _ := os.ReadDir("/home/public/")
		// if err {
		// 	fmt.Printf(err)
		// }
		var filenames []string
		for _, file := range files {
			filenames = append(filenames, file.Name())
		}
		// filenames := files.Name()
		// fmt.Printf(filename)
		tpl.ExecuteTemplate(w, "download.html", filenames)
		IPAddress := r.Header.Get("X-Real-Ip")
        	if IPAddress == "" {
                	IPAddress = r.Header.Get("X-Forwarded-For")
        	}
        	if IPAddress == "" {
                	IPAddress = r.RemoteAddr
        	}
        	log.Println(net.ParseIP(IPAddress))
		return
	}
	myFile := r.FormValue("myFile")
	myPassword := r.FormValue("myPassword")
	log.Printf("myFile: %s\n", myFile)
	log.Printf("myPassword: %s\n", myPassword)
	IPAddress := r.Header.Get("X-Real-Ip")
    	if IPAddress == "" {
        	IPAddress = r.Header.Get("X-Forwarded-For")
    	}
    	if IPAddress == "" {
        	IPAddress = r.RemoteAddr
    	}
	log.Println(net.ParseIP(IPAddress))
	decryptFile(myFile , myPassword)
	w.Header().Set("Content-Disposition", "attachment; filename=" + myFile)
	http.ServeFile(w, r, "/home/public/decrypted" + myFile)
	err := os.Remove("/home/public/decrypted" + myFile)
	if err != nil {
		log.Println(err)
		return
	}
}

func directory(w http.ResponseWriter, r *http.Request) {
	files, err := ioutil.ReadDir("home/public/")
	if err != nil {
        	log.Println(err)
    	}
	for _, file := range files {
        	fmt.Fprintf(w, file.Name() + "\n")
    	}
}

func uploadFile(w http.ResponseWriter, r *http.Request) {
	log.Println("r.method:", r.Method)
	// if method is GET then load form, if not then upload successfull message
	if r.Method == "GET" {
		tpl.ExecuteTemplate(w, "upload.html", nil)
		return
	}
	// func (r *Request) ParseMultipartForm(maxMemory int64) error
	r.ParseMultipartForm(10)
	// func (r *Request) FormFile(key string) (multipart.File, *multipart.FileHeader, error)
	file, fileHeader, err := r.FormFile("myFile")
	myPassword := r.FormValue("myPassword")
	
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()
	log.Printf("fileHeader.Filename: %v\n", fileHeader.Filename)
	log.Printf("fileHeader.Size: %v\n", fileHeader.Size)
	log.Printf("fileHeader.Header: %v\n", fileHeader.Header)
	log.Printf("myPassword: %s\n", myPassword)

	// tempFile, err := ioutil.TempFile("images", "upload-*.png")
	contentType := fileHeader.Header["Content-Type"][0]
	log.Println("Content Type:", contentType)
	var osFile *os.File
	// func TempFile(dir, pattern string) (f *os.File, err error)
	/*
	if contentType == "image/jpeg" {
		osFile, err = ioutil.TempFile("images", "*.jpg")
	} else if contentType == "application/pdf" {
		osFile, err = ioutil.TempFile("PDFs", "*.pdf")
	} else if contentType == "text/javascript" {
		osFile, err = ioutil.TempFile("js", "*.js")
	}
	*/
	osFile, err = ioutil.TempFile("public", "*" + fileHeader.Filename)
	log.Println("error:", err)
	defer osFile.Close()

	// func ReadAll(r io.Reader) ([]byte, error)
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
	}
	// func (f *File) Write(b []byte) (n int, err error)

	osFile.Write(encrypt(fileBytes, myPassword))
	fmt.Fprintf(w, "Your File was Successfully Uploaded!\n")
	IPAddress := r.Header.Get("X-Real-Ip")
        if IPAddress == "" {
                IPAddress = r.Header.Get("X-Forwarded-For")
        }
        if IPAddress == "" {
                IPAddress = r.RemoteAddr
        }
        log.Println(net.ParseIP(IPAddress))
	//w.Header().Set("Content-Disposition", "attachment; filename=book.pdf")
	//http.ServeFile(w, r, "./public/book.pdf")
}
/*
// Recursively get all file paths in directory, including sub-directories.
func GetAllFilePathsInDirectory(dirpath string) ([]string, error) {
      var paths []string
      err := filepath.Walk(dirpath, func(path string, info os.FileInfo, err error) error {
              if err != nil {
                      return err
              }
              if !info.IsDir() {
                      paths = append(paths, path)
              }
              return nil
      })
      if err != nil {
              return nil, err
      }

      return paths, nil
}

// Recursively parse all files in directory, including sub-directories.
func ParseDirectory(dirpath string) (*template.Template, error) {
      paths, err := GetAllFilePathsInDirectory(dirpath)
      if err != nil {
              return nil, err
      }
      return template.ParseFiles(paths...)
}
*/
