package dashboard

import (
	"github.com/BruteForceFencer/core/hitcounter"
	"github.com/BruteForceFencer/core/version"
	"html/template"
	"net/http"
	"os"
	"path"
	"path/filepath"
)

// installPath is the path to the BFF installation in slash form.
var installPath string

func init() {
	determineInstallPath()
}

// HandleAssets serves the HTML, CSS and JS assets for the dashboard.
func (s *Server) HandleAssets(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		s.serveHomePage(w, r)
	} else {
		assetPath := path.Join("assets", r.URL.Path)
		assetPath = filepath.FromSlash(fromInstallPath(assetPath))
		http.ServeFile(w, r, assetPath)
	}
}

func (s *Server) serveHomePage(w http.ResponseWriter, r *http.Request) {
	data := struct {
		ListenAddress string
		ListenType    string
		Version       string
		Directions    []hitcounter.Direction
	}{
		ListenAddress: s.conf.ListenAddress,
		ListenType:    s.conf.ListenType,
		Version:       version.Version,
		Directions:    s.conf.Directions,
	}

	htmlPath := filepath.FromSlash(fromInstallPath("assets/dashboard.html"))
	t, err := template.ParseFiles(htmlPath)
	if err != nil {
		http.Error(w, "Unable to find server files.", http.StatusInternalServerError)
		return
	}

	t.Execute(w, data)
}

// fromInstallPath returns the absolute path by appending p to the installation
// path.  The result is in slash form.
func fromInstallPath(p string) string {
	return path.Join(installPath, p)
}

// determineInstallPath sets the installPath variable.
func determineInstallPath() {
	// The path to the executable binary.
	var execPath string

	// Argument 0 converted to slash notation.
	arg0 := filepath.ToSlash(path.Clean(os.Args[0]))
	if len(arg0) > 0 && arg0[0] == '/' {
		// The 0th arg is the absolute path.
		execPath = path.Clean(arg0)
	} else {
		wd, _ := os.Getwd()
		execPath = path.Join(wd, arg0)
	}

	installPath = path.Dir(path.Dir(execPath))
}
