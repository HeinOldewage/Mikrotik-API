package Mikrotik

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"strings"
)

type Router struct {
	connection      net.Conn
	response        map[string]chan Sentence
	defaultresponse chan Sentence
	idList          chan string
}

func New() *Router {
	maxId := 100
	res := &Router{
		nil,
		make(map[string]chan Sentence),
		make(chan Sentence),
		make(chan string, maxId),
	}
	for k := 0; k < maxId; k++ {
		res.idList <- fmt.Sprint(k)
	}
	return res
}

func (r *Router) Connect(addr string) error {
	var err error
	r.connection, err = net.Dial("tcp", addr)
	if err == nil {

		go func() {
			for {
				//Get length, maximal 5 bytes
				sentence := &Sentence{}
				var id string = ""
				sentenceDone := false
				for !sentenceDone {
					var lengthbytes []byte = make([]byte, 5)
					var length int
					for k := 0; k < 5; k++ {
						_, err := r.connection.Read(lengthbytes[k : k+1])
						if err != nil {
							//Panic??
							//How do we report this error?
							log.Println(err)
							return
						}
						if length, _, err = decodelength(lengthbytes[:k+1]); err == nil {
							lengthbytes = lengthbytes[:k+1]
							break
						}
					}
					if length == 0 {
						sentenceDone = true
						continue
					}

					var wordbytes []byte = make([]byte, length)
					for length > 0 {
						var n int
						n, err = r.connection.Read(wordbytes[len(wordbytes)-length:])
						length -= n
					}
					if err != nil {
						//Panic??
						//How do we report this error?
						panic(err)
					}
					word, bl, err := Decode(append(lengthbytes, wordbytes...))
					if (err != nil) || (bl != (len(lengthbytes) + len(wordbytes))) {
						//Panic??
						//How do we report this error?
						panic(err)
					}

					if strings.Index(string(word), ".tag=") == 0 {
						id = string(word[len(".tag="):])
					}
					sentence.Add(word)
				}
				if len(*sentence) == 0 {
					log.Println("len(*sentence) == 0 ")
					continue
				}
				if id != "" {

					r.response[id] <- *sentence
				} else {
				}
				if (sentence.Contains(Word("!done")) /*|| sentence.Contains(Word("!trap"))*/) && (id != "") {
					close(r.response[id])
					r.idList <- id
				}
			}
		}()

	}
	return err
}

func (r *Router) Login(user, pass string) error {
	login := make(Sentence, 0)
	login.Add(Command("/login"))
	c, _, err := r.SendSentence(login)
	if err != nil {
		return err
	}
	chalSen := <-c
	if !chalSen.Contains(Word("=ret=")) {
		return fmt.Errorf("Response did not have challenge %v", []Word(chalSen))
	}
	retW := chalSen[chalSen.Index("=ret=")]
	chal, err := hex.DecodeString(strings.Split(string(retW[1:]), "=")[1])
	if err != nil {
		return err
	}
	Sum := md5.New()
	hashBytes := append([]byte{0}, []byte(pass)...)
	hashBytes = append(hashBytes, []byte(chal)...)

	_, err = Sum.Write(hashBytes)
	if err != nil {
		return err
	}
	resp := Sum.Sum(nil)
	login.Add(Attribute("name", user))
	login.Add(Attribute("response", "00"+hex.EncodeToString(resp)))
	c, _, err = r.SendSentence(login)
	if err != nil {
		return err
	}
	chalSen = <-c
	if chalSen.Contains(Word("!trap")) {
		return fmt.Errorf("%v", []Word(chalSen))
	}

	return nil
}

func (r *Router) SendSentence(mess Sentence) (response chan Sentence, tag string, err error) {
	if r.connection == nil {
		return nil, "", fmt.Errorf("Router not connected")
	}
	tag = <-r.idList
	response = make(chan Sentence)
	mess.Add(APIAttribute("tag", tag))
	r.response[tag] = response
	_, err = r.connection.Write(mess.Encode())

	return response, tag, err
}
