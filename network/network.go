package network

import(
  "net"
	//"fmt"
)

func GetLocalIPAddr() (string, error) {

  ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

  for _, i := range ifaces {
      //fmt.Printf("Iface #%d: %v\n", n, i)
      addrs, _ := i.Addrs()
		  if err != nil {
				return "", err
			}
      for _, addr := range addrs {
          var ip net.IP
          switch v := addr.(type) {
          case *net.IPNet:
                  ip = v.IP
          case *net.IPAddr:
                  ip = v.IP
          }
          //fmt.Printf("  - IP:  %v\n", ip)
					if !ip.IsLoopback() {
						return ip.String(), nil
					}
		}
	}

	 err = networkError{s:"network.GetLocalIPAddr(): Could not find network interface"}
	return "", err
}

type networkError struct {
	s string
}

func (e networkError) Error() string {
	return e.s
}
