package lib

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

var version = "1.13.3"

type Config struct {
	User           string
	TweetID        string
	OutputDir      string `default:"/Downloads"`
	MediaType      string `default:"all"`
	NumberOfTweets int    `default:"100"`
	Single         string
	Videos         bool   `default:"false"`
	Images         bool   `default:"false"`
	UrlOnly        bool   `default:"false"`
	Retweets       bool   `default:"false"`
	RetweetOnly    bool   `default:"false"`
	Size           string `default:"orig"`
	Update         bool   `default:"true"`
	Format         string `default:"{DATE} {USERNAME} {NAME} {TITLE} {ID}"`
	Datefmt        string `default:"2006-01-02"`
	Login          string `default:"false"`
	Loginp         string `default:"false"`
	Twofa          bool   `default:"false"`
	Proxy          string `default:""`
	Nologo         bool   `default:"false"`
	Printversion   bool   `default:"false"`
}

func quitWithError(flags *flag.FlagSet, err string) {
	fmt.Fprintln(os.Stderr, err)
	flags.Usage()
	os.Exit(1)
}

func Configure() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.User, "user", "", "User you want to download")
	flag.StringVar(&cfg.Single, "tweet", "", "Single tweet to download")
	flag.IntVar(&cfg.NumberOfTweets, "N", 0, "Number of tweets to download")
	flag.BoolVar(&cfg.Images, "img", false, "Download images only")
	flag.BoolVar(&cfg.Videos, "video", false, "Download videos only")

	var videosImages bool
	flag.BoolVar(&videosImages, "all", false, "Download images and videos")
	flag.BoolVar(&cfg.Retweets, "retweet", false, "Download retweet too")
	flag.BoolVar(&cfg.UrlOnly, "url", false, "Return media URL without downloading it")
	flag.BoolVar(&cfg.RetweetOnly, "retweet-only", false, "Download only retweets")
	flag.StringVar(&cfg.Size, "size", "large", "Choose size between small|normal|large (default large)")
	flag.BoolVar(&cfg.Update, "update", false, "Download missing tweets only")
	flag.StringVar(&cfg.OutputDir, "output", "", "Output directory")
	flag.StringVar(&cfg.Format, "file-format", "", "Formatted name for the downloaded file, {DATE} {USERNAME} {NAME} {TITLE} {ID}")
	flag.StringVar(&cfg.Datefmt, "date-format", "", "Apply custom date format. (https://go.dev/src/time/format.go)")
	flag.StringVar(&cfg.Login, "login", "", "Login (needed for NSFW tweets)")
	flag.StringVar(&cfg.Loginp, "login-plaintext", "", "Plain text login (needed for NSFW tweets)")
	flag.BoolVar(&cfg.Twofa, "2fa", false, "Use 2fa")
	flag.StringVar(&cfg.Proxy, "proxy", "", "Use proxy (proto://ip:port)")
	flag.BoolVar(&cfg.Printversion, "version", false, "Print version and exit")

	// Custom usage message
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "twmd: Apiless twitter media downloader\n\nUsage:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  twmd -u Spraytrains -o ~/Downloads -a -r -n 300\n")
		fmt.Fprintf(os.Stderr, "  twmd -u Spraytrains -o ~/Downloads -R -U -n 300\n")
		fmt.Fprintf(os.Stderr, "  twmd --proxy socks5://127.0.0.1:9050 -t 156170319961391104\n")
		fmt.Fprintf(os.Stderr, "  twmd -t 156170319961391104\n")
		fmt.Fprintf(os.Stderr, "  twmd -t 156170319961391104 -f \"{DATE} {ID}\"\n")
		fmt.Fprintf(os.Stderr, "  twmd -t 156170319961391104 -f \"{DATE} {ID}\" -d \"2006-01-02_15-04-05\"\n")
	}

	flag.Parse()

	if cfg.Printversion {
		fmt.Println("version:", version)
		os.Exit(1)
	}

	if videosImages {
		cfg.Videos = true
		cfg.Images = true
	}

	if cfg.User == "" && cfg.Single == "" {
		quitWithError(flag.CommandLine, "You must specify a user (-user) or a tweet (-tweet)")
	}

	if !cfg.Videos && !cfg.Images && cfg.Single == "" {
		quitWithError(flag.CommandLine, "You must specify what to download. (-img) for images, (-video) for videos or (-all) for both")
	}

	var re = regexp.MustCompile(`{ID}|{DATE}|{NAME}|{USERNAME}|{TITLE}`)
	if cfg.Format != "" && !re.MatchString(cfg.Format) {
		quitWithError(flag.CommandLine, "Invalid format specified. Must include at least one of {ID}, {DATE}, {NAME}, {USERNAME}, or {TITLE}")
		flag.Usage()
		os.Exit(1)
	}

	re = regexp.MustCompile("small|normal|large")
	if !re.MatchString(cfg.Size) {
		quitWithError(flag.CommandLine, "Error in size: Must be one of small, normal, large")
		os.Exit(1)
	}

	cfg.OutputDir = filepath.Join(cfg.OutputDir, cfg.User)
	if cfg.Videos {
		os.MkdirAll(filepath.Join(cfg.OutputDir, "video"), os.ModePerm)
	}

	if cfg.Images {
		os.MkdirAll(filepath.Join(cfg.OutputDir, "img"), os.ModePerm)
	}

	return cfg
}
