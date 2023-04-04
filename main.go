package main

import (
	"log"
	"os"
	"github.com/joho/godotenv"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"context"
	"github.com/redis/go-redis/v9"
	"strconv"
	"fmt"
	"strings"
    "net/http"
    "io/ioutil"
    "time"
)

func main(){
  env := getEnv("env","local")
  
  log.Print("starting engine")
  err := godotenv.Load()
  if err != nil {
    log.Fatal("Error loading .env file")
  }
    log.Println("env file loaded!")

  // check mysql
  mysql_enabled := getEnv("MYSQL_ENABLED","false")
  if (mysql_enabled=="true"){
	  mysql_host := getEnv("MYSQL_HOST","localhost")
	  mysql_port := getEnv("MYSQL_PORT","3306")
	  mysql_user := getEnv("MYSQL_USER","root")
	  mysql_pass := getEnv("MYSQL_PASS","")
	  mysql_db   := getEnv("MYSQL_DB","")

	  db, err := sql.Open("mysql", mysql_user+":"+mysql_pass+"@tcp("+mysql_host+":"+mysql_port+")/"+mysql_db)
	  _, err = db.Query("SHOW DATABASES")
	  if err != nil {
		  log.Println("MYSQL_NOT_WORKING",err)
		  alertToSlack("MYSQL NOT WORKING",env)
	  }else{
	  	  log.Println("MYSQL_WORKING")
	  	  db.Close()
	  }
	}

  //check mongo
  mongo_enabled := getEnv("MONGO_ENABLED","false")
  if (mongo_enabled=="true"){
		mongo_uri := getEnv("MONGO_URI","localhost")
		client, _ := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongo_uri))
		if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		    log.Println("MONGO_NOT_WORKING",err)
		    alertToSlack("MONGO NOT WORKING",env)
		}else{
			log.Println("MONGO_WORKING")
		}

	}

	//check redis
	redis_enabled := getEnv("REDIS_ENABLED","false")
	if (redis_enabled=="true"){
		redis_host := getEnv("REDIS_HOST","127.0.0.1")
		redis_port := getEnv("REDIS_PORT","6379")
		redis_pass := getEnv("REDIS_PASS","")
		redis_db,_ := strconv.Atoi(getEnv("REDIS_DB","0"))

		rdb := redis.NewClient(&redis.Options{
	        Addr:     redis_host+":"+redis_port,
	        Password: redis_pass, 
	        DB:       redis_db,
	    })
	    res := fmt.Sprintf("%s",rdb.Ping(context.TODO()));
	   	if ( res == "ping: PONG" ){
			log.Println("REDIS_WORKING")
	   	}else{
			log.Println("REDIS_NOT_WORKING",res)
			alertToSlack("REDIS NOT WORKING",env)
	   	}
	}


}

func getEnv(key, fallback string) string {
    if value, ok := os.LookupEnv(key); ok {
        return value
    }
    return fallback
}

func alertToSlack(msg string,env string){
	currentTime := time.Now()
	currnetDateString := fmt.Sprintf("%d-%d-%d %d:%d:%d\n",currentTime.Year(),currentTime.Month(),currentTime.Day(),currentTime.Hour(),currentTime.Hour(),currentTime.Second())
	url := getEnv("SLACK_WEBHOOK","")
	method := "POST"

	payload := strings.NewReader(`{
	    "blocks": [
	        {
	            "type": "section",
	            "text": {
	                "type": "mrkdwn",
	                "text": "`+currnetDateString+`  :bangbang::bangbang::bangbang:  `+env+" => "+msg+`  :bangbang::bangbang::bangbang:"
	            }
	        }
	    ]
	}`)

	client := &http.Client {
	}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
		if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))
}