// all the handlers that reflects the functionalities that an admin can perform
package cont

import (
	"encoding/json"
	"fmt"
	Res "main/Response"
	"main/db"
	"main/models"
	"main/utils"
	"net/http"
	"os"

	"github.com/bogem/id3v2"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"golang.org/x/crypto/bcrypt"
)




func Admin_login(w http.ResponseWriter,r *http.Request){

	utils.SetHeader(w)

	if r.Method != http.MethodPost {
		// w.WriteHeader(http.StatusMethodNotAllowed)
		Res.Response("Method Not Allowed ",405,"use correct http method","",w)
		return
	}

	input:=make(map[string]string)
	json.NewDecoder(r.Body).Decode(&input)
	err := validation.Validate(input,
		validation.Map(
			// song_path cannot be empty
			
			validation.Key("email",validation.Required,is.Email),
			validation.Key("password",validation.Required),
			
		),
	)
	
	if err!=nil{

		Res.Response("Bad Request",400,err.Error(),"",w)
		return
	}


	var admin models.User

	er:=db.DB.Where("email=?",input["email"]).Find(&admin).Error
	if er!=nil{

		Res.Response("Server error",500,er.Error(),"",w)
	}
	
	er1 := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(input["password"]))
	if er1!=nil{

		Res.Response("Unathorized",401,"Invalid Password","",w)
		return
	}else{

		//give admin access cookie

		claims:=models.Claims{Role: "admin"}
		adminToken:=GenerateNewToken(&claims)
		cookie:=&http.Cookie{ Name:"token",Value: adminToken}

		http.SetCookie(w,cookie)

		Res.Response("OK",200,"admin login success","",w)
	}


}











// @Description Add Song into app
// @Accept json
// @Produce json
// @Param  details body string true "enter PATH of song SchemaExample({"path":"/home/chicmic/Downloads/"})
// @Tags Admin
// @Success 200 {object} models.Response
// @Router /add-song [post]
func Add_Song(w http.ResponseWriter,r *http.Request) {

	utils.SetHeader(w)

	if r.Method != http.MethodPost {
		// w.WriteHeader(http.StatusMethodNotAllowed)
		Res.Response("Method Not Allowed ",405,"use correct http method","",w)
		return
	}
	

	//takes audiofile path from r.body

	var pathh models.Path

	input:=make(map[string]string)
	json.NewDecoder(r.Body).Decode(&input)
	err := validation.Validate(input,
		validation.Map(
			// song_path cannot be empty
			
			validation.Key("path",validation.Required),
			
		),
	)
	
	if err!=nil{

		Res.Response("Bad Request",400,err.Error(),"",w)
		return
	}
	pathh.Path=input["path"]

	

	 //Open the audio file
	
	 file, err := os.Open(pathh.Path)
	 if err != nil {
		fmt.Println("err in file opening ")
		//  log.Fatal(err)
		Res.Response("Bad Request",400,"Provide proper audio file path","",w)
		return
		 
	 }
	 defer file.Close()
 
	 tag, err := id3v2.ParseReader(file,id3v2.Options{Parse: true})
	 if err != nil {
		//  log.Fatal(err)
		fmt.Println("err",err)
		 Res.Response("Server error",500,"error in audio_file parsing ","",w)
		 return

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
	 er:=db.DB.Create(&audiofile).Error
	 if er != nil {
		 fmt.Println(er.Error())
		 Res.Response("Bad Request",400,er.Error(),"",w)
		 return
	 }
	

	 Res.Response("Success",200,"Audio file saved to database","",w)
	fmt.Println("Audio file saved to database")
 

}


// @Description Add Thumbnail for Song 
// @Accept json
// @Produce json
// @Param  details body string true "enter Song id and path of thumbnail of song SchemaExample({"id":"xyz","img_path":"/"})
// @Tags Admin
// @Success 200 {object} models.Response
// @Router /add-img [post]
func Add_Thumbnail_Img(w http.ResponseWriter,r * http.Request){


	//take input audio_file id (in which you want to add IMg)
	//and img path
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		
		Res.Response("Method Not Allowed ",405,"use correct http method","",w)
		
	}
	

	input:=make(map[string]string)
	json.NewDecoder(r.Body).Decode(&input)
	err := validation.Validate(input,
		validation.Map(
			// id and img_path cannot be empty
			validation.Key("id", validation.Required),
			validation.Key("img_path",validation.Required),
			
		),
	)
	
	if err!=nil{

		Res.Response("Bad Request",400,err.Error(),"",w)
		return
	}
	
	var song models.AudioFile
	

	

	song.Img_Path=input["img_path"]
	song.ID=input["id"]
	query:="select exists(select * from audio_files where id='"+song.ID+"' and img_path='"+song.Img_Path+"');"
	var exists bool
	db.DB.Raw(query).Scan(&exists)
	if exists{
		Res.Response("Bad Request",400,"already exists","",w)
		return 
	}	
	er:=db.DB.Where("id=?",song.ID).Updates(&song).Error
	if er!=nil{

		Res.Response("server error",500,er.Error(),"",w)
	}

	
	Res.Response("OK",200,"Thumbnail added successfully","",w)

}


// @Description Create Album
// @Accept json
// @Produce json
// @Param  details body string true "enter Song id and album name SchemaExample({"song_id":"xyz","album_name":"name"})
// @Tags Admin
// @Success 200 {object} models.Response
// @Router /create-album [post]
func Create_Album(w http.ResponseWriter,r * http.Request){


	//take the input song_id,album_name
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		
		Res.Response("Method Not Allowed ",405,"use correct http method","",w)
		
	}

	

	input:=make(map[string]string)
	json.NewDecoder(r.Body).Decode(&input)
	err := validation.Validate(input,
		validation.Map(
			
			validation.Key("album_name", validation.Required),
			validation.Key("song_id",validation.Required),
			
			
		),
	)
	
	if err!=nil{

		Res.Response("Bad Request",400,err.Error(),"",w)
		return
	}

	var album models.Album
	album.Album_name=input["album_name"]
	album.Song_id=input["song_id"]

	query:="select exists(select * from albums where song_id='"+album.Song_id+"' and album_name='"+album.Album_name+"');"
	var exists bool
	db.DB.Raw(query).Scan(&exists)
	if exists{
		Res.Response("Bad Request",400,"already exists","",w)
		return 
	}

	er:=db.DB.Create(&album).Error
	if er!=nil{

		Res.Response("server error",500,er.Error(),"",w)
		
	}

	
	Res.Response("OK",200,"Album created","",w)



}

