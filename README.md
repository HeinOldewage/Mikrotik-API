# Mikrotik-API
To download this repo use

~~~~
go get github.com/HeinOldewage/Mikrotik-API
~~~~


An example usage

~~~~
	router := Mikrotik.New()

	err := router.Connect("ipAdress:8728")
	if err != nil {
		log.Fatal(err)
	}

	err = router.Login("username", "password")
	if err != nil {
		log.Fatal(err)
	}

	sen := make(Mikrotik.Sentence, 0)
	sen.Add(Mikrotik.Command("/interface/wireless/registration-table/listen"))

	ch, tag, err := router.SendSentence(sen)

	if err != nil {
		log.Fatal(err)
	}

	for res := range ch {
		log.Println(res)
	}
~~~~

	
Note that this API supports commands that have continuous output (like ```listen``` and ```ping```). 
Multiple commands can also be run concurrently. This is facilitated by using tags. 

To stop a long running commmand do the following, where tag is the tag is returned from ```SendSentence```.
	
	
~~~~
sen = make(Mikrotik.Sentence, 0)
sen.Add(Mikrotik.Command("/cancel"))
sen.Add(Mikrotik.Attribute("tag", tag))
_, _, err := router.SendSentence(sen)
~~~~

