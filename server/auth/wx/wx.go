package wx

type Server struct {
	AppID     string
	AppSecret string
}

//func (s *Server) Resolve(code string) (string, error) {
//	resp, err := weapp.Login(s.AppID, s.AppSecret, code)
//	if err != nil {
//		return "", err
//	}
//	if err := resp.GetResponseError(); err != nil {
//		return "", err
//	}
//
//	return resp.OpenID, nil
//}

func (s *Server) Resolve(code string) (string, error) {
	return "wxbjafiodjasif132132", nil
}
