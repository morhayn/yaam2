package artifact

// func validate(requestURI string) error {
// log.Tracef("requestURI: '%s'", requestURI)
// dir := filepath.Dir(requestURI)
// ext := filepath.Ext(requestURI)
// if dir == "/" || ext == "" {
// return fmt.Errorf("requestURI: '%s' should start with a: '/' and contain an extension", requestURI)
// }

// regex := `^/([a-z]+)/([0-9a-z-]+)/`
// re := regexp.MustCompile(regex)
// repoTypeAndName := re.FindStringSubmatch(requestURI)
// if len(repoTypeAndName) <= 2 {
// return fmt.Errorf("no repo type or name detected. Verify whether the regex: '%s' matches the URL: '%s'", regex, requestURI)
// }
// repoType := repoTypeAndName[1]
// repoName := repoTypeAndName[2]

// if err := allowedRepo(repoName, repoType); err != nil {
// return err
// }

// return nil
// }

// func allowedRepo(name, artifactType string) error {
// repos := viper.GetStringSlice("publications." + artifactType)
// if !slices.Contains(repos, name) {
// return fmt.Errorf("repository: '%s' is not allowed. Allowed repos: '%v'", name, repos)
// }

// return nil
// }
