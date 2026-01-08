package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

var (
	jellyfin  string
	token     string
	userID    string
	imageSize string
)

type JellyfinItem struct {
	Id              string            `json:"Id"`
	Name            string            `json:"Name"`
	Type            string            `json:"Type"`
	SeriesId        string            `json:"SeriesId"`
	ImageTags       map[string]string `json:"ImageTags"`
	CommunityRating float64           `json:"CommunityRating"`
	OfficialRating  string            `json:"OfficialRating"`
	PremiereDate    string            `json:"PremiereDate"`
}

type Card struct {
	Title         string  `json:"title"`
	Subtitle      string  `json:"subtitle"`
	Image         string  `json:"image"`
	Href          string  `json:"href"`
	Rating        float64 `json:"rating,omitempty"`
	Year          string  `json:"year,omitempty"`
	ContentRating string  `json:"contentRating,omitempty"`
}

type cacheEntry struct {
	Data []Card
	Time time.Time
}

var cache = map[string]cacheEntry{}
var mu sync.Mutex

func fetchLatest(filter string) ([]Card, error) {
	mu.Lock()
	if c, ok := cache[filter]; ok && time.Since(c.Time) < 60*time.Second {
		mu.Unlock()
		return c.Data, nil
	}
	mu.Unlock()

	userID := os.Getenv("JELLYFIN_USERID")
	fmt.Println("userID:", userID)

	if userID == "" {
		return nil, fmt.Errorf("JELLYFIN_USERID environment variable is not set")
	}

	url := jellyfin + "/Users/" + userID + "/Items/Latest?Limit=20"
	if filter != "" {
		url += "&IncludeItemTypes=" + filter
	}

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-Emby-Token", token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var items []JellyfinItem
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, err
	}

	out := make([]Card, 0, len(items))
	for _, i := range items {
		imageID := i.Id

		// Episode → Series poster fallback
		if i.Type == "Episode" && i.SeriesId != "" {
			imageID = i.SeriesId
		}

		year := ""
		if len(i.PremiereDate) >= 4 {
			year = i.PremiereDate[:4]
		}

		out = append(out, Card{
			Title:         i.Name,
			Subtitle:      i.Type,
			Image:         jellyfin + "/Items/" + imageID + "/Images/Primary" + imageSize,
			Href:          jellyfin + "/web/index.html#!/details?id=" + i.Id,
			Rating:        i.CommunityRating,
			Year:          year,
			ContentRating: i.OfficialRating,
		})
	}

	mu.Lock()
	cache[filter] = cacheEntry{Data: out, Time: time.Now()}
	mu.Unlock()

	return out, nil
}

func apiHandler(filter string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := fetchLatest(filter)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		json.NewEncoder(w).Encode(map[string][]Card{
			"items": data,
		})
	}
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Recently Added</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }

        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
            background: #0f0f0f;
            color: #fff;
            padding: 20px;
            overflow-x: hidden;
        }

        .header {
            margin-bottom: 20px;
            display: flex;
            justify-content: space-between;
            align-items: center;
            flex-wrap: wrap;
            gap: 15px;
        }

        h2 {
            font-size: 24px;
            font-weight: 600;
            color: #fff;
        }

        .filter-tabs {
            display: flex;
            gap: 10px;
            flex-wrap: wrap;
        }

        .filter-btn {
            background: rgba(255, 255, 255, 0.1);
            border: none;
            color: #fff;
            padding: 8px 16px;
            border-radius: 6px;
            cursor: pointer;
            font-size: 14px;
            transition: all 0.2s;
        }

        .filter-btn:hover {
            background: rgba(255, 255, 255, 0.15);
        }

        .filter-btn.active {
            background: #00a4dc;
        }

        .carousel {
            display: flex;
            gap: 16px;
            overflow-x: auto;
            overflow-y: hidden;
            scroll-behavior: smooth;
            padding-bottom: 20px;
            scrollbar-width: thin;
            scrollbar-color: rgba(255, 255, 255, 0.3) transparent;
        }

        .carousel::-webkit-scrollbar { height: 8px; }
        .carousel::-webkit-scrollbar-track {
            background: rgba(255, 255, 255, 0.05);
            border-radius: 4px;
        }
        .carousel::-webkit-scrollbar-thumb {
            background: rgba(255, 255, 255, 0.3);
            border-radius: 4px;
        }

        .card {
            flex: 0 0 auto;
            width: 159px;
            cursor: pointer;
            transition: transform 0.3s;
            text-decoration: none;
        }

        .card:hover { transform: scale(1.05); }

        .card-image {
            width: 100%;
            height: 220px;
            border-radius: 8px;
            overflow: hidden;
            position: relative;
            background: linear-gradient(135deg, #1a1a1a 0%, #2a2a2a 100%);
        }

        .card-image img {
            width: 100%;
            height: 100%;
            object-fit: cover;
        }

        .card-overlay {
            position: absolute;
            bottom: 0;
            left: 0;
            right: 0;
            background: linear-gradient(to top, rgba(0,0,0,0.95) 0%, transparent 100%);
            padding: 12px;
            opacity: 0;
            transition: opacity 0.3s;
        }

        .card:hover .card-overlay { opacity: 1; }

        .rating-info {
            display: flex;
            align-items: center;
            gap: 8px;
            flex-wrap: wrap;
        }

        .rating {
            display: flex;
            align-items: center;
            gap: 4px;
            font-size: 14px;
            font-weight: 600;
            color: #ffd700;
        }

        .year {
            font-size: 12px;
            color: #999;
        }

        .content-rating {
            background: rgba(255, 255, 255, 0.2);
            padding: 2px 6px;
            border-radius: 3px;
            font-size: 11px;
            font-weight: 600;
        }

        .card-info { margin-top: 10px; }

        .card-title {
            font-size: 14px;
            font-weight: 500;
            color: #fff;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
            margin-bottom: 4px;
        }

        .card-subtitle {
            font-size: 12px;
            color: #999;
        }

        .loading-state, .error-state {
            text-align: center;
            padding: 40px;
            color: #999;
        }

        .error-state { color: #e50914; }

        @keyframes pulse {
            0%, 100% { opacity: 1; }
            50% { opacity: 0.5; }
        }

        .skeleton { animation: pulse 1.5s ease-in-out infinite; }
    </style>
</head>
<body>
    <div class="header">
        <h2>Recently Added</h2>
        <div class="filter-tabs">
            <button class="filter-btn active" data-filter="">All</button>
            <button class="filter-btn" data-filter="movies">Movies</button>
            <button class="filter-btn" data-filter="tv">TV Shows</button>
            <button class="filter-btn" data-filter="music">Music</button>
            <button class="filter-btn" data-filter="books">Books</button>
        </div>
    </div>

    <div class="carousel" id="carousel">
        <div class="loading-state">Loading...</div>
    </div>

    <script>
        let currentFilter = '';

        async function loadContent(filter) {
            const carousel = document.getElementById('carousel');
            carousel.innerHTML = '<div class="loading-state">Loading...</div>';

            try {
                const endpoint = filter ? '/latest/' + filter : '/latest';
                const response = await fetch(endpoint);
                if (!response.ok) throw new Error('Failed to fetch');

                const data = await response.json();
                carousel.innerHTML = '';

                if (!data.items || data.items.length === 0) {
                    carousel.innerHTML = '<div class="loading-state">No items found</div>';
                    return;
                }

                data.items.forEach(item => {
                    const card = document.createElement('a');
                    card.className = 'card';
                    card.href = item.href;
                    card.target = '_blank';

                    let overlayContent = '';
                    if (item.rating || item.year || item.contentRating) {
                        const parts = [];
                        if (item.rating) {
                            parts.push('<div class="rating">★ ' + item.rating.toFixed(1) + '</div>');
                        }
                        if (item.year) {
                            parts.push('<span class="year">' + item.year + '</span>');
                        }
                        if (item.contentRating) {
                            parts.push('<span class="content-rating">' + item.contentRating + '</span>');
                        }
                        overlayContent = '<div class="card-overlay"><div class="rating-info">' +
                            parts.join('') + '</div></div>';
                    }

                    card.innerHTML =
                        '<div class="card-image">' +
                            '<img src="' + item.image + '" alt="' + item.title + '" ' +
                                 'onerror="this.style.display=\'none\'">' +
                            overlayContent +
                        '</div>' +
                        '<div class="card-info">' +
                            '<div class="card-title">' + item.title + '</div>' +
                            '<div class="card-subtitle">' + item.subtitle + '</div>' +
                        '</div>';

                    carousel.appendChild(card);
                });

            } catch (error) {
                console.error('Error:', error);
                carousel.innerHTML = '<div class="error-state">Failed to load content</div>';
            }
        }

        document.querySelectorAll('.filter-btn').forEach(btn => {
            btn.addEventListener('click', () => {
                document.querySelectorAll('.filter-btn').forEach(b => b.classList.remove('active'));
                btn.classList.add('active');
                currentFilter = btn.dataset.filter;
                loadContent(currentFilter);
            });
        });

        loadContent(currentFilter);
        setInterval(() => loadContent(currentFilter), 300000);
    </script>
</body>
</html>`))
}

func main() {
	jellyfin = os.Getenv("JELLYFIN_URL")
	token = os.Getenv("JELLYFIN_TOKEN")
	userID = os.Getenv("JELLYFIN_USERID")
	imageSize = os.Getenv("POSTER_IMAGE_SIZE")

	if jellyfin == "" || token == "" || userID == "" {
		log.Fatal("Missing required environment variables")
	}

	// JSON API endpoints (existing)
	http.HandleFunc("/latest", apiHandler(""))
	http.HandleFunc("/latest/movies", apiHandler("Movie"))
	http.HandleFunc("/latest/tv", apiHandler("Series"))
	http.HandleFunc("/latest/music", apiHandler("MusicAlbum"))
	http.HandleFunc("/latest/books", apiHandler("Book"))

	// Dashboard HTML endpoint (new)
	http.HandleFunc("/", dashboardHandler)
	http.HandleFunc("/dashboard", dashboardHandler)

	log.Println("Jellyfin Dashboard running on :7654")
	log.Println("  Dashboard: http://localhost:7654/")
	log.Println("  JSON API:  http://localhost:7654/latest")
	log.Fatal(http.ListenAndServe(":7654", nil))
}
