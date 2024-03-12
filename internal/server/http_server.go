package server

import (
	"embed"
	"fmt"
	koolo "github.com/hectorgimenez/koolo/internal"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/helper"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
)

type HttpServer struct {
	logger  *slog.Logger
	manager *koolo.SupervisorManager
}

var (
	//go:embed all:assets
	assets embed.FS
	//go:embed all:templates
	templates embed.FS

	configTpl       = template.Must(template.ParseFS(templates, "templates/config.html"))
	indexTpl        = template.Must(template.ParseFS(templates, "templates/index.html"))
	charSettingsTpl = template.Must(template.ParseFS(templates, "templates/character_settings.html"))
)

func New(logger *slog.Logger, manager *koolo.SupervisorManager) *HttpServer {
	return &HttpServer{
		logger:  logger,
		manager: manager,
	}
}

func (s *HttpServer) Listen(port int) error {
	http.HandleFunc("/", s.getRoot)
	http.HandleFunc("/config", s.config)
	http.HandleFunc("/addCharacter", s.add)
	http.HandleFunc("/editCharacter", s.edit)
	http.HandleFunc("/start", s.startSupervisor)
	http.HandleFunc("/stop", s.stopSupervisor)
	http.HandleFunc("/togglePause", s.togglePause)

	assets, _ := fs.Sub(assets, "assets")
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.FS(assets))))

	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func (s *HttpServer) getRoot(w http.ResponseWriter, r *http.Request) {
	if !helper.HasAdminPermission() {
		tmpl := template.Must(template.ParseFS(templates, "templates/admin_required.html"))
		tmpl.Execute(w, nil)
		return
	}

	if config.Koolo.FirstRun {
		http.Redirect(w, r, "/config", http.StatusSeeOther)
		return
	}

	s.index(w)
}

func (s *HttpServer) startSupervisor(w http.ResponseWriter, r *http.Request) {
	s.manager.Start(r.URL.Query().Get("characterName"))
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *HttpServer) stopSupervisor(w http.ResponseWriter, r *http.Request) {
	s.manager.Stop(r.URL.Query().Get("characterName"))
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *HttpServer) togglePause(w http.ResponseWriter, r *http.Request) {
	s.manager.TogglePause(r.URL.Query().Get("characterName"))
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *HttpServer) index(w http.ResponseWriter) {
	status := make(map[string]koolo.Stats)
	for _, supervisorName := range s.manager.AvailableSupervisors() {
		status[supervisorName] = koolo.Stats{
			SupervisorStatus: koolo.NotStarted,
		}

		status[supervisorName] = s.manager.Status(supervisorName)
	}

	indexTpl.Execute(w, IndexData{Status: status})
}

func (s *HttpServer) config(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			configTpl.Execute(w, ConfigData{KooloCfg: config.Koolo, ErrorMessage: "Error parsing form"})
			return
		}

		newConfig := *config.Koolo
		newConfig.D2RPath = r.Form.Get("d2rpath")
		newConfig.D2LoDPath = r.Form.Get("d2lodpath")
		newConfig.FirstRun = false

		err = config.ValidateAndSaveConfig(newConfig)
		if err == nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		configTpl.Execute(w, ConfigData{KooloCfg: config.Koolo, ErrorMessage: err.Error()})
		return
	}

	configTpl.Execute(w, ConfigData{KooloCfg: config.Koolo, ErrorMessage: ""})
}

func (s *HttpServer) add(w http.ResponseWriter, r *http.Request) {
	var err error
	if r.Method == "POST" {
		err = r.ParseForm()
		if err != nil {

		}

		supervisorName := r.Form.Get("name")
		err = config.CreateFromTemplate(supervisorName)
		if err != nil {
			charSettingsTpl.Execute(w, CharacterSettings{
				ErrorMessage: err.Error(),
				Supervisor:   supervisorName,
			})

			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	supervisor := r.URL.Query().Get("supervisor")
	cfg := config.Characters["template"]
	if supervisor != "" {
		cfg = config.Characters[supervisor]
	}

	charSettingsTpl.Execute(w, CharacterSettings{
		Supervisor:   supervisor,
		CharacterCfg: cfg,
	})
}

func (s *HttpServer) edit(w http.ResponseWriter, r *http.Request) {

}
