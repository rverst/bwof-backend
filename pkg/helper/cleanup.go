package helper

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"io"
)

func Close(c io.Closer, s string) {
	err := c.Close()
	if err != nil {
		log.Error().Err(err).Msg(fmt.Sprintf("helper.Close | %s", s))
	}
}

func CloseRC(c io.ReadCloser, s string) {
  err := c.Close()
  if err != nil {
    log.Error().Err(err).Msg(fmt.Sprintf("helper.CloseRC | %s", s))
  }
}
