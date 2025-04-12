package main

import (
	"fmt"

	"encoding/json"

	"github.com/google/uuid"
)

type Movie struct {
	MovieId          string `json:"movieId" dynamodbav:"movieId"`
	Title            string `json:"title" dynamodbav:"title"`
	ReleaseYear      uint16 `json:"releaseYear" dynamodbav:"releaseYear"`
	Genre            string `json:"genre" dynamodbav:"genre"`
	CoverUrl         string `json:"coverUrl" dynamodbav:"coverUrl"`
	GeneratedSummary string `json:"generatedSummary,omitempty" dynamodbav:"generatedSummary,omitempty"`
}

var movies = []Movie{
	{Title: "Pulp Fiction", ReleaseYear: 1994, Genre: "Crime, Drama", CoverUrl: "https://movies-api-data.s3.ap-south-1.amazonaws.com/images/pulp-fiction-1994.jpg"},
	{Title: "The Matrix", ReleaseYear: 1999, Genre: "Science Fiction, Action", CoverUrl: "https://movies-api-data.s3.ap-south-1.amazonaws.com/images/the-matrix-1999.jpg"},
	{Title: "Forrest Gump", ReleaseYear: 1994, Genre: "Drama, Romance", CoverUrl: "https://movies-api-data.s3.ap-south-1.amazonaws.com/images/forrest-gump-1994.jpg"},
	{Title: "The Godfather", ReleaseYear: 1972, Genre: "Crime, Drama", CoverUrl: "https://movies-api-data.s3.ap-south-1.amazonaws.com/images/the-godfather-1972.jpg"},
	{Title: "Interstellar", ReleaseYear: 2014, Genre: "Science Fiction, Adventure", CoverUrl: "https://movies-api-data.s3.ap-south-1.amazonaws.com/images/interstellar-2014.jpg"},
	{Title: "Titanic", ReleaseYear: 1997, Genre: "Romance, Drama", CoverUrl: "https://movies-api-data.s3.ap-south-1.amazonaws.com/images/titanic-1997.jpg"},
	{Title: "Jurassic Park", ReleaseYear: 1993, Genre: "Science Fiction, Adventure", CoverUrl: "https://movies-api-data.s3.ap-south-1.amazonaws.com/images/jurassic-park-1993.jpg"},
	{Title: "The Lion King", ReleaseYear: 1994, Genre: "Animation, Adventure", CoverUrl: "https://movies-api-data.s3.ap-south-1.amazonaws.com/images/the-lion-king-1994.jpg"},
	{Title: "Fight Club", ReleaseYear: 1999, Genre: "Drama, Thriller", CoverUrl: "https://movies-api-data.s3.ap-south-1.amazonaws.com/images/fight-club-1999.jpg"},
	{Title: "Avatar", ReleaseYear: 2009, Genre: "Science Fiction, Action", CoverUrl: "https://movies-api-data.s3.ap-south-1.amazonaws.com/images/avatar-2009.jpg"},
	{Title: "The Empire Strikes Back", ReleaseYear: 1980, Genre: "Science Fiction, Action", CoverUrl: "https://movies-api-data.s3.ap-south-1.amazonaws.com/images/the-empire-strikes-back-1980.jpg"},
	{Title: "Schindler's List", ReleaseYear: 1993, Genre: "Drama, History", CoverUrl: "https://movies-api-data.s3.ap-south-1.amazonaws.com/images/schindlers-list-1993.jpg"},
	{Title: "The Lord of the Rings: The Fellowship of the Ring", ReleaseYear: 2001, Genre: "Fantasy, Adventure", CoverUrl: "https://movies-api-data.s3.ap-south-1.amazonaws.com/images/lotr-the-fellowship-of-rings-2001.jpg"},
	{Title: "Gladiator", ReleaseYear: 2000, Genre: "Action, Drama", CoverUrl: "https://movies-api-data.s3.ap-south-1.amazonaws.com/images/gladiator-2000.jpg"},
	{Title: "The Silence of the Lambs", ReleaseYear: 1991, Genre: "Thriller, Crime", CoverUrl: "https://movies-api-data.s3.ap-south-1.amazonaws.com/images/the-silence-of-the-lambs-1991.jpg"},
	{Title: "Back to the Future", ReleaseYear: 1985, Genre: "Science Fiction, Adventure", CoverUrl: "https://movies-api-data.s3.ap-south-1.amazonaws.com/images/back-to-the-future-1985.jpg"},
	{Title: "Parasite", ReleaseYear: 2019, Genre: "Thriller, Drama", CoverUrl: "https://movies-api-data.s3.ap-south-1.amazonaws.com/images/parasite-2019.jpg"},
	{Title: "Mad Max: Fury Road", ReleaseYear: 2015, Genre: "Action, Science Fiction", CoverUrl: "https://movies-api-data.s3.ap-south-1.amazonaws.com/images/max-mad-fury-road-2015.jpg"},
	{Title: "The Avengers", ReleaseYear: 2012, Genre: "Action, Superhero", CoverUrl: "https://movies-api-data.s3.ap-south-1.amazonaws.com/images/the-avengers-2012.jpg"},
	{Title: "Good Will Hunting", ReleaseYear: 1997, Genre: "Drama", CoverUrl: "https://movies-api-data.s3.ap-south-1.amazonaws.com/images/good-will-hunting-1997.jpg"},
}

const (
	AWS_REGION  string = "ap-south-1"
	BUCKET_NAME string = "movies-api-data"
	TABLE_NAME  string = "Movies"
	MODEL_ID    string = "anthropic.claude-3-sonnet-20240229-v1:0"
)

func main() {

	for index := range movies {
		id, _ := uuid.NewV7()
		movies[index].MovieId = fmt.Sprintf("%v", id)
	}
	// fmt.Println(movies)

	_, err := json.Marshal(movies)
	if err != nil {
		fmt.Println(err)
	}
	// fmt.Println(string(jsonD))

	// Insert into dynamodb
	// if err := PutItems_DynamoDB(movies); err != nil {
	// 	fmt.Println(err)
	// }
	// if err := GetMovies(); err != nil {
	// 	fmt.Println(err)
	// }
	// if err := GetMoviesByYear(2025); err != nil {
	// 	fmt.Println(err)
	// }
	// if err := GetMovieSummary("0195ea79-8284-772c-a501-61f2a6369726"); err != nil {
	// 	fmt.Println(err)
	// }
	// if err := UpdateMovie_DB("0195ea79-8284-772c-a501-61f2a6369726", "1994", "test2"); err != nil {
	// 	fmt.Println(err)
	// }
	// if err := generateSummary("Provide a short summary of 100 words for the movie 'Forrest Gump', released in 1994, which falls under the genre Drama, Romance."); err != nil {
	// 	fmt.Println(err)
	// }
	// if err := GetMovieById("01956766-a4a2-7836-bd37-0c1cb0ac1f3d"); err != nil {
	// 	fmt.Println(err)
	// }
	if err := DeleteMovieById("01956766-a4a2-7836-bd37-0c1cb0ac1f3d"); err != nil {
		fmt.Println(err)
	}
}
