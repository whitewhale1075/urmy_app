package urmy_app

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/unrolled/render"
	"github.com/urfave/negroni"
	urmy_handler "github.com/whitewhale1075/urmy_handler"
)

type AppHandler struct {
	http.Handler
	db         urmy_handler.DBHandler
	rc         urmy_handler.JWTHandler
	sj         urmy_handler.SaJuHandler
	sa         urmy_handler.SaJuAnalyzer
	configfile configFile
}

type Success struct {
	Success bool `json:"success"`
}

type configFile struct {
	serverkey string
}

type AuthUser struct {
	LoginID         string `json:"loginId"`
	Password        string `json:"password"`
	UserdataExist   bool   `json:"userdataexist"`
	ServerdataExist bool   `json:"serverdataexist"`
}

type User struct {
	LoginID   string `json:"loginId"`
	Password  string `json:"password"`
	Nickname  string `json:"nickname"`
	Name      string `json:"name"`
	Birthdate string `json:"birthdate"`
	PhoneNo   string `json:"phoneNo"`
	Gender    bool   `json:"gender"`
}

type Friends struct {
	Identifier  string
	GivenName   string
	FamilyName  string
	PhonesLabel string
	PhonesValue string
}

type SajuResult struct {
	LoginId     string
	Grade       string
	Description string
}

type Cookie struct {
	Name       string
	Value      string
	Path       string
	Domain     string
	Expires    time.Time
	RawExpires string
	MaxAge     int
	Secure     bool
	HttpOnly   bool
	Raw        string
	Unparsed   []string
}

type PersonSaju struct {
	LoginID   string `json:"LoginId"`
	YearChun  string `json:"YearChun"`
	YearJi    string `json:"YearJi"`
	MonthChun string `json:"MonthChun"`
	MonthJi   string `json:"MonthJi"`
	DayChun   string `json:"DayChun"`
	DayJi     string `json:"DayJi"`
	TimeChun  string `json:"TimeChun"`
	TimeJi    string `json:"TimeJi"`
	DaeunChun string `json:"DaeunChun"`
	DaeUnJi   string `json:"DaeUnJi"`
	SaeunChun string `json:"SaeunChun"`
	SaeunJi   string `json:"SaeunJi"`
}

var rd *render.Render = render.New()

func MakeHandler() *AppHandler {
	r := mux.NewRouter()
	n := negroni.New(
		negroni.NewRecovery(),
		negroni.NewLogger(),
		//		negroni.HandlerFunc(CheckSignin),
		//		negroni.NewStatic(http.Dir("public"))
	)
	n.UseHandler(r)
	f, err := ioutil.ReadFile("/etc/config/configfile.json")
	if err != nil {
		fmt.Println(err)
		return nil
	}

	var configfile configFile
	json.Unmarshal(f, &configfile)

	a := &AppHandler{
		Handler:    n,
		db:         urmy_handler.NewDBHandler(),
		rc:         urmy_handler.NewJWTHandler(),
		sj:         urmy_handler.NewSaJuHandler(),
		sa:         urmy_handler.NewSaJuAnalyzer(),
		configfile: configfile,
	}

	r.HandleFunc("/login", a.getUrMyUserHandler).Methods("POST")
	r.HandleFunc("/register", a.addUrMyUserHandler).Methods("POST")
	r.HandleFunc("/registeradditional", a.addUrMyAdditionalHandler).Methods("POST")
	r.HandleFunc("/friendlist", a.friendlistHandler).Methods("POST")
	return a
}

func (a *AppHandler) friendlistHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	var t []Friends
	err = json.Unmarshal(body, &t)
	if err != nil {
		panic(err)
	}
	userid := a.tokenVerifyProcessHandler(w, r)

	mysaju, errmysaju := a.db.GetMySaju(userid)
	if errmysaju != nil {
		fmt.Println(errmysaju)
	}
	sjtable := a.sj.GetSajuTable()
	satable := a.sa.GetAnalyzerTable()
	//friendssaju := make([]PersonSaju, len(t))
	mysajuanalyzed := make([]urmy_handler.Person, len(t))
	mygoonghabevaluated := make([]SajuResult, len(t))
	friendsajuanalyzed := make([]urmy_handler.Person, len(t))
	friendgoonghabevaluated := make([]SajuResult, len(t))
	var count int = 0
	for i := 0; i < len(t); i++ {
		//phones := strings.ReplaceAll(t[i].PhoneNo[0].Value, "-", "")
		result, err := a.db.GetUrMyFriendList(t[i].PhonesValue)
		if err != nil {
			fmt.Println(err)
		} else {
			mysajuanalyzed[count], friendsajuanalyzed[count] = a.sa.Find_GoongHab(mysaju, result, sjtable.Chungan, sjtable.Jiji, satable.Sibsung, satable.Sib2Unsung)
			mygoonghabevaluated[count].Grade, friendgoonghabevaluated[count].Grade, mygoonghabevaluated[count].Description, friendgoonghabevaluated[count].Description = a.sa.Evaluate_GoonbHab(mysajuanalyzed[count], friendsajuanalyzed[count])
			mygoonghabevaluated[count].LoginId = result.LoginID

			count++
		}
	}

	rd.JSON(w, http.StatusOK, mygoonghabevaluated)
}

func (a *AppHandler) getUrMyUserHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	//temp := string(body[:])
	var t AuthUser
	err = json.Unmarshal(body, &t)
	if err != nil {
		panic(err)
	}

	//sess, err := a.gs.SessionStart(w, r)
	//defer sess.SessionRelease(w)
	//sess.SessionRelease(w)

	if t.ServerdataExist == false {

		http.SetCookie(w, &http.Cookie{
			Name:  "firebasekey",
			Value: a.configfile.serverkey,
		})
	}

	result, t2 := a.db.GetUrMyUser(t.LoginID, t.Password, t.UserdataExist)
	if result {
		accessjwtgen, err := a.rc.GernerateAccessJWT(t.LoginID)
		if err != nil {
			panic(err)
		}
		refreshjwtgen, err := a.rc.GernerateRefreshJWT(t.LoginID)
		if err != nil {
			panic(err)
		}
		http.SetCookie(w, &http.Cookie{
			Name:  "accesstoken",
			Value: accessjwtgen.AccessToken,
		})
		http.SetCookie(w, &http.Cookie{
			Name:  "refreshtoken",
			Value: refreshjwtgen.RefreshToken,
		})
		a.rc.CreateAccessAuth(t.LoginID, accessjwtgen)
		a.rc.CreateRefreshAuth(t.LoginID, refreshjwtgen)
		if t.UserdataExist {
			rd.JSON(w, http.StatusOK, Success{true})
		} else {
			rd.JSON(w, http.StatusOK, &t2)
		}
	} else {
		rd.JSON(w, http.StatusOK, Success{false})
	}
}

func (a *AppHandler) addUrMyUserHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	//temp := string(body[:])
	var t User
	err = json.Unmarshal(body, &t)
	if err != nil {
		panic(err)
	}
	result, err := a.db.AddUrMyUser(t.LoginID, t.Password, t.Nickname, t.Name, t.PhoneNo, t.Gender, t.Birthdate)
	if result != nil {
		palja := a.sj.ExtractSaju(result)
		palja = a.sj.ExtractDaeUnSaeUn(result, palja)
		sajuerr := a.db.InputUrMySaJuInfo(t.LoginID, palja)
		if sajuerr != nil {
			rd.JSON(w, http.StatusCreated, Success{true})
		} else {
			rd.JSON(w, http.StatusCreated, Success{false})
		}
		//a.db.InputUrMySaJuDaeSaeUnInfo(t.LoginID, palja)
	} else {
		rd.JSON(w, http.StatusCreated, Success{false})
	}
}

func (a *AppHandler) addUrMyAdditionalHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	//temp := string(body[:])
	var t User
	err = json.Unmarshal(body, &t)
	if err != nil {
		panic(err)
	}

	result, err := a.db.AddUrMyAdditionalInfo(t.LoginID, t.Birthdate)
	if result != nil {
		palja := a.sj.ExtractSaju(result)
		palja = a.sj.ExtractDaeUnSaeUn(result, palja)
		a.db.InputUrMySaJuInfo(t.LoginID, palja)

		rd.JSON(w, http.StatusCreated, Success{true})
	} else {
		rd.JSON(w, http.StatusCreated, Success{false})
	}
}

func (a *AppHandler) Close() {
	a.db.Close()
}

func (a *AppHandler) tokenVerifyProcessHandler(w http.ResponseWriter, r *http.Request) string {
	at, aterr := a.rc.ExtractAccessTokenMetadata(r)
	if aterr != nil {
		fmt.Println("ExtractAccessTokenMetadata2")
		fmt.Println(aterr)
		//rt, rterr := a.rc.ExtractRefreshTokenMetadata(r)
	}
	rt, rterr := a.rc.ExtractRefreshTokenMetadata(r)
	if rterr != nil {
		fmt.Println("ExtractAccessTokenMetadata2")
		fmt.Println(rterr)
	}
	vat, vaterr := a.rc.FetchAccessAuth(at)
	if vaterr != nil {
		fmt.Println("FetchAccessAuth")
		fmt.Println(vaterr)
	}
	vrt, vrterr := a.rc.FetchRefreshAuth(rt)
	if vrterr != nil {
		fmt.Println("FetchRefreshAuth")
		fmt.Println(vrt)
	}
	return vat
}
