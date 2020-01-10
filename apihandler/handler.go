package apihandler

import (
	pb "Cw_authService/proto"
	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"time"
)

const (
	secret = "transavro"
)

type AuthService struct {
	*mongo.Collection
}

// Login For this example, the credentials are stored in
// in-memory in variabe uname and only works for 1 user.
// In a realworld example, a database or some store would
// be used for storing the username and hashed password.
func (s *AuthService) Login(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {

	log.Println("Authorizing user", req.GetUname())
	if req.GetUname() == "" || req.GetPwd() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "missing uname or password")
	}

	findRaw, err := s.Collection.FindOne(ctx, bson.D{{"name", req.Uname}}).DecodeBytes()
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "No user Found.")
	}
	//checking password.
	if err := bcrypt.CompareHashAndPassword(findRaw.Lookup("password").Value, []byte(req.GetPwd())); err != nil {
		log.Println("auth failed")
		return nil, status.Error(codes.PermissionDenied, "auth failed")
	}

	// create jwt token
	// see reserved claims https://tools.ietf.org/html/rfc7519#section-4.1
	// see jwt example here https://godoc.org/github.com/dgrijalva/jwt-go#example-New--Hmac
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			"exp":  time.Now().Add(time.Minute * 20).Unix(),
			"iss":  "authservice",
			"aud":  "user",
			"name": req.Uname,
		},
	)

	// this example uses a simple string secret. You can also
	// use JWT package to specify an RSA public cert here as well.
	tokenString, err := token.SignedString([]byte(secret))

	if err != nil {
		log.Println(err)
		return nil, status.Error(codes.Internal, "internal login problem")
	}

	log.Printf("User %s logged in OK, JWT token: %s\n", req.Uname, tokenString)
	return &pb.AuthResponse{Token: tokenString}, nil
}
