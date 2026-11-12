package faction

import "testing"

func TestNamesRoundTrip(t *testing.T) {
	for _, p := range []Player{None, CentralPowers, Entente} {
		got, err := ByName(p.String())
		if err != nil || got != p {
			t.Errorf("ByName(%q) = %v, %v", p.String(), got, err)
		}
	}
	if _, err := ByName("Neutral"); err == nil {
		t.Error("an unknown name was accepted")
	}
	if s := Player(7).String(); s != "Player(7)" {
		t.Errorf("unknown player prints %q", s)
	}
}

func TestOpponentAndOf(t *testing.T) {
	if CentralPowers.Opponent() != Entente || Entente.Opponent() != CentralPowers || None.Opponent() != None {
		t.Error("wrong opponents")
	}
	for country, want := range map[string]Player{"Germany": CentralPowers, "Ottoman": CentralPowers, "France": Entente, "Britain": Entente} {
		if got := Of(country); got != want {
			t.Errorf("Of(%s) = %v, want %v", country, got, want)
		}
	}
}
