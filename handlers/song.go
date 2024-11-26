package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/schema"
)
type SongRequest struct {
  Group string `schema:"group"`
  Song  string `schema:"song"`
}
type SongDetail struct {
  ReleaseDate string `json:"releaseDate"`
  Text        string `json:"text"`
  Link        string `json:"link"`
}

func GetSongInfo(w http.ResponseWriter, r *http.Request) {
	group := r.URL.Query().Get("group")
song := r.URL.Query().Get("song")
fmt.Printf("Received query params - group: %s, song: %s\n", group, song)
	
  var request SongRequest
  decoder := schema.NewDecoder()
  if err := decoder.Decode(&request, r.URL.Query()); err != nil {
    http.Error(w, "Failed to decode query parameters", http.StatusBadRequest)
    return
  }
  if request.Group == "" || request.Song == "" {
    http.Error(w, "Missing 'group' or 'song' parameter", http.StatusBadRequest)
    return
  }
  songDetails := SongDetail{
    ReleaseDate: "16.07.2006",
    Text: `I can dim the lights and sing you songs
Full of sad things
We can do the tango, just for two
I can serenade and gently play
On your heart strings
Be a Valentino, just for you

"Ooh love, ooh lover boy
What're you doing tonight?"
Set my alarm, turn on my charm
That's because I'm a good old-fashioned lover boy

Ooh, let me feel
Your heartbeat (Grow faster, faster)
Ooh, can you feel my love heat? (Ohh)
Come on and sit on my hot seat of love
And tell me how do you feel, right after all
I'd like for you and I to go romancing
Say the word, your wish is my command

"Ooh love, ooh lover boy
What're you doing tonight? Hey boy"
Write my letter, feel much better
I'll use my fancy patter on the telephone
See upcoming rock shows
Get tickets for your favorite artists
You might also like
Take Your Mask Off
Tyler, The Creator
St. Chroma
Tyler, The Creator
Love Language
SZA
[Bridge: Freddie Mercury, Mike Stone]
When I'm not with you, think of you always
I miss you (I miss those long hot summer nights)
When I'm not with you, think of me always
Love you, love you
Hey boy where do you get it from?
Hey boy where did you go?
I learned my passion in the good old-fashioned
School of lover boys

Dining at the Ritz, we'll meet at nine
(One, two, three, four, five, six, seven, eight, nine o'clock) precisely
I will pay the bill, you taste the wine
Driving back in style in my saloon will do quite nicely
Just take me back to yours, that will be fine (Come on and get it)

Ooh love (There he goes again)
Ooh lover boy (Who's my good Old-fashioned lover boy?)
(Ooh ooh)
What're you doing tonight? Hey boy!
Everything's all right, just hold on tight
That's because I'm a good old
Fashioned (Fashioned) lover boy`,
    Link: "https://www.youtube.com/watch?v=Xsp3_a-PMTw",
  }
  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusOK)
  if err := json.NewEncoder(w).Encode(songDetails); err != nil {
    http.Error(w, "Unable to encode song details", http.StatusInternalServerError)
  }
}