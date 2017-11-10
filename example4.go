package main

import (
        "io"  
        "bufio"
        "net"
        "bytes"
	"errors"
        "fmt"
	"flag"
	"log"
	"net/http"
	"strings"
        "path"
	groupcache "./groupcache"
)

var Store = map[string][]byte{
	"red":   []byte("#FF0000"),
	"green": []byte("#00FF00"),
	"blue":  []byte("#0000FF"),
}

var Group = groupcache.NewGroup("foobar", 64<<20, groupcache.GetterFunc(
	func(ctx groupcache.Context, key string, dest groupcache.Sink) error {
		log.Println("looking up", key)
		v, ok := Store[key]
		if !ok {
			return errors.New("color not found")
		}
		dest.SetBytes(v)
		return nil
	},
))

var( 
        addr     string // The address of our httpd server.
        daddr    string // The address of the doozer server.
        dictName string // The name of the dictionary.
        dictAddr string // The address of the dictionary server.
        // This is our groupcache stuff.
        pool *groupcache.HTTPPool
        dict *groupcache.Group
)


func init() {
        flag.StringVar(&addr, "addr", "127.0.0.1:8000",
                "the addr:port on which this server should run.")
        flag.StringVar(&daddr, "doozer", "127.0.0.1:8046",
                "the addr:port on which doozer is running.")
        flag.StringVar(&dictName, "dictname", "gcide",
                "the name of the dictionary to query.")
        flag.StringVar(&dictAddr, "dictaddr", "dict.org:2628",
                "the addr:port to the dict server to query.")
}

func main() {
	//addr = flag.String("addr", ":8080", "server address")
	peers := flag.String("pool", "http://localhost:8080", "server pool list")
	flag.Parse()

	p := strings.Split(*peers, ",")
	pool := groupcache.NewHTTPPool(p[0])
	pool.Set(p...)

        dict = groupcache.NewGroup("dict", 64<<20, groupcache.GetterFunc(
                func(ctx groupcache.Context, key string, dest groupcache.Sink) error {
                        def, err := query(key)
                        if err != nil {
                                err = fmt.Errorf("querying remote dictionary: %v", err)
                                log.Println(err)
                                return err 
                        }   

                        log.Println("retrieved remote definition for", key)
                        dest.SetString(def)
                        return nil 
                })) 


	http.Handle("/define/", http.HandlerFunc(handler)) 
	log.Println(http.ListenAndServe(addr, nil))
}

// handler handles all incoming requests for a definition.
func handler(w http.ResponseWriter, r *http.Request) {
        log.Println("received request:", r.Method, r.URL.Path)
        word := strings.Trim(path.Base(r.URL.Path), "/")

        // Get the definition from groupcache and write it out.
        var data []byte
        err := dict.Get(nil, word, groupcache.AllocatingByteSliceSink(&data))
        if err != nil {
                log.Println("retreiving definition for", word, "-", err)
                w.WriteHeader(http.StatusInternalServerError)
                return
        }
        io.Copy(w, bytes.NewReader(data))
}

// query is a helper function for the groupcache that queries a remote
// dict server for the first definition of the given word.
func query(word string) (string, error) {
        // NOTE: I am aware this is brittle and doesn't really follow the
        // protocol all that well.
        conn, err := net.Dial("tcp", dictAddr)
        if err != nil {
                return "", fmt.Errorf("connecting to dict: %v", err)
        }
        defer conn.Close()

        // Send the DEFINE request and read the response into a buffer.
        fmt.Fprintf(conn, "DEFINE %s %s\r\n", dictName, word)
        scanner := bufio.NewScanner(conn)
        var response bytes.Buffer
        for scanner.Scan() {
                // Read the line, trim any excess new lines
                line := scanner.Text()
                line = strings.Trim(line, "\r\n")
                if strings.HasPrefix(line, "2") || strings.HasPrefix(line, "1") {
                        // Skip over any control data.
                        continue
                }
                if line == "." || line == "" {
                        // Quit when we reach the end of the first definition.
                        break
                }

                // Store the line we just read.
                response.WriteString(line)
                response.WriteString("\n")
        }

        // Check for errors after the scan.
        if err := scanner.Err(); err != nil {
                return "", fmt.Errorf("reading line from connection: %v", err)
        }

        // Send the QUIT message and return the definition.
        fmt.Fprintf(conn, "QUIT\r\n")
        return response.String(), nil
}
                         
