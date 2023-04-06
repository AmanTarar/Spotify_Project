// all the handlers that reflects the functionalities that an admin can perform
package cont

import (
	"encoding/json"
	"fmt"
	"log"
	"main/db"
	"main/models"
	con "main/utils"
	"net/http"
	"os"
	"time"

	"github.com/bogem/id3v2"
	"github.com/golang-jwt/jwt/v4"
)


func Create_Admin(){


	var admin models.User

	admin.Role="admin"
	admin.Name="aman-admin"
	db.DB.Create(&admin)
	
}
func GetToken(){

	// jwt authentication token
	
	expirationTime := time.Now().Add(365* 24 * time.Hour)
	fmt.Println("expiration time is: ", expirationTime)

	// check if the user is valid then only create token

	var user models.User
	db.DB.Where("role=?", "admin").First(&user)
	claims := models.Claims{

		Role:user.Role,
		User_id:user.User_id,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	// fmt.Println("claims: ", claims)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	fmt.Println("token: ", token)
	tokenString, err := token.SignedString((con.Jwt_key))
	if err != nil {
		fmt.Println("error is :", err)
		// w.WriteHeader(http.StatusInternalServerError)
		
	}
	// fmt.Println("tokenString",tokenString)
	user.Token=tokenString
	db.DB.Where("role=?", "admin").Updates(&user)


	
}


func Add_Song(w http.ResponseWriter,r *http.Request) {


	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	//takes audiofile path from r.body

	var pathh models.Path
	json.NewDecoder(r.Body).Decode(&pathh)

	fmt.Println("path in post req",pathh.Path)

	 //Open the audio file
	
	 file, err := os.Open(pathh.Path)
	 if err != nil {
		fmt.Println("err in file opening ")
		 log.Fatal(err)
		 
	 }
	 defer file.Close()
 
	 tag, err := id3v2.ParseReader(file,id3v2.Options{Parse: true})
	 if err != nil {
		 log.Fatal(err)
	 }
	// Create a new AudioFile object
	var audiofile models.AudioFile

	audiofile.Path=pathh.Path

	audiofile.Name=tag.Title()
	audiofile.Artist=tag.Artist()
	
	
	//calculate the size of the audiofile
	fileinfo,err:=file.Stat()
	
	audiofileinBytes:=fileinfo.Size()

	audiofile.Size=float64(audiofileinBytes/(1024*1024))

	
	

 
	 // Create a new record in the database
	 result := db.DB.Create(&audiofile)
	 if result.Error != nil {
		 fmt.Println(result.Error)
		 return
	 }
	fmt.Fprint(w,"Audio file saved to database")
	fmt.Println("Audio file saved to database")
 

}



func Add_Thumbnail_Img(w http.ResponseWriter,r * http.Request){


	//take input audio_file id (in which you want to add IMg)
	//pass the img path in params 

	var song models.AudioFile

	json.NewDecoder(r.Body).Decode(&song)

	image_Path:=r.URL.Query().Get("img_path")

	song.Img_Path=image_Path

	db.DB.Where("id=?",song.ID).Updates(&song)

	fmt.Fprint(w,"Thumbnail added successfully")

}

func Create_Album(w http.ResponseWriter,r * http.Request){


	//take the input song_id,album_name

	var album models.Album

	json.NewDecoder(r.Body).Decode(&album)

	db.DB.Create(&album)

	fmt.Fprint(w,"Album created")


}

