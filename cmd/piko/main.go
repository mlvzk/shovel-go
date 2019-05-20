package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/mlvzk/piko"
	"github.com/mlvzk/piko/service"
	"github.com/mlvzk/qtils/commandparser"
	"github.com/mlvzk/qtils/commandparser/commandhelper"
	"gopkg.in/cheggaaa/pb.v1"
)

var (
	discoveryMode bool
	stdoutMode    bool
	formatStr     string
	targets       []string
	userOptions   = map[string]string{}
)

func handleArgv(argv []string) {
	parser := commandparser.New()
	helper := commandhelper.New()

	helper.SetName("piko")
	helper.SetVersion("alpha")
	helper.AddAuthor("mlvzk")

	helper.AddUsage(
		"piko [urls...]",
		"piko 'https://www.youtube.com/watch?v=dQw4w9WgXcQ'",
		"piko 'https://www.youtube.com/watch?v=dQw4w9WgXcQ' -stdout | mpv -",
	)

	parser.AddOption(helper.EatOption(
		commandhelper.NewOption("help").Alias("h").Boolean().Description("Prints this page").Build(),
		commandhelper.
			NewOption("format").
			Alias("f").
			Description("File path format, ex: -format %[id].%[ext]. Use %[default] to fill with default format, ex: downloads/%[default]").
			Build(),
		commandhelper.
			NewOption("option").
			Alias("o").
			Arrayed().
			Description("Download options, ex: --option quality=best").
			Build(),
		commandhelper.
			NewOption("discover").
			Alias("d").
			Boolean().
			Description("Discovery mode, doesn't download anything, only outputs information").
			Build(),
		commandhelper.
			NewOption("stdout").
			Boolean().
			Description("Output download media to stdout").
			Build(),
	)...)

	cmd, err := parser.Parse(argv)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	if cmd.Booleans["help"] {
		fmt.Print(helper.Help())
		os.Exit(1)
	}

	cmd.Args = helper.FillDefaults(cmd.Args)
	errs := helper.Verify(cmd.Args, cmd.Arrayed)
	for _, err := range errs {
		log.Println(err)
	}
	if len(errs) != 0 {
		os.Exit(1)
	}

	formatStr = cmd.Args["format"]
	discoveryMode = cmd.Booleans["discover"]
	stdoutMode = cmd.Booleans["stdout"]

	for _, option := range cmd.Arrayed["option"] {
		keyValue := strings.Split(option, "=")
		key, value := keyValue[0], keyValue[1]

		userOptions[key] = value
	}

	targets = cmd.Positionals
}

func main() {
	handleArgv(os.Args)

	services := piko.GetAllServices()

	// target = "https://boards.4channel.org/adv/thread/20765545/i-want-to-be-the-very-best-like-no-one-ever-was"
	// target = "https://imgur.com/t/article13/EfY6CxU"
	// target = "https://www.youtube.com/watch?v=Gs069dndIYk"
	// target = "https://www.youtube.com/watch?v=7IwYakbxmxo"
	// target = "https://www.instagram.com/p/Bv9MJCsAvZV/"
	// target = "https://soundcloud.com/musicpromouser/mac-miller-ok-ft-tyler-the-creator"
	// target = "https://twitter.com/deadprogram/status/1090554988768698368"
	// target = "https://www.facebook.com/groups/veryblessedimages/permalink/478153699389793/"

	for _, target := range targets {
		if target == "" {
			log.Println("Target can't be empty")
			break
		}

		var foundAnyService bool
		for _, s := range services {
			if !s.IsValidTarget(target) {
				continue
			}

			foundAnyService = true
			log.Printf("Found valid service: %s\n", reflect.TypeOf(s).Name())
			iterator := s.FetchItems(target)

			for !iterator.HasEnded() {
				items, err := iterator.Next()
				if err != nil {
					log.Printf("Iteration error: %v; target: %v\n", err, target)
					break
				}

				for _, item := range items {
					handleItem(s, item)
				}
			}

			break
		}
		if !foundAnyService {
			log.Fatalf("Couldn't find a valid service for url '%s'. Your link is probably unsupported.\n", target)
		}
	}
}

func handleItem(s service.Service, item service.Item) {
	if discoveryMode {
		log.Println("Item:\n" + prettyPrintItem(item))
		return
	}

	options := mergeStringMaps(item.DefaultOptions, userOptions)

	reader, err := s.Download(item.Meta, options)
	if err != nil {
		log.Printf("Download error: %v, item: %+v\n", err, item)
		return
	}
	defer tryClose(reader)

	if stdoutMode {
		io.Copy(os.Stdout, reader)
		return
	}

	nameFormat := item.DefaultName
	if formatStr != "" {
		nameFormat = strings.Replace(formatStr, "%[default]", item.DefaultName, -1)
	}
	name := format(nameFormat, item.Meta)

	if sizedIO, ok := reader.(service.Sized); ok {
		bar := pb.New64(int64(sizedIO.Size())).SetUnits(pb.U_BYTES)
		bar.Prefix(truncateString(name, 25))
		bar.Start()
		defer bar.Finish()

		reader = bar.NewProxyReader(reader)
	}

	file, err := os.Create(name)
	if err != nil {
		log.Printf("Error creating file: %v, name: %v\n", err, name)
		return
	}
	defer file.Close()

	_, err = io.Copy(file, reader)
	if err != nil {
		log.Printf("Error copying from source to file: %v, item: %+v", err, item)
		return
	}
}
