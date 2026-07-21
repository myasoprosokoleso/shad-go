//go:build !solution

package hotelbusiness

import "slices"

type Guest struct {
	CheckInDate  int
	CheckOutDate int
}

type Load struct {
	StartDate  int
	GuestCount int
}

func ComputeLoad(guests []Guest) []Load {
	type event struct {
		date  int
		delta int
	}

	n := len(guests)
	events := make([]event, 2*n)
	for i, g := range guests {
		events[2*i] = event{date: g.CheckInDate, delta: 1}
		events[2*i+1] = event{date: g.CheckOutDate, delta: -1}
	}

	slices.SortFunc(events, func(a, b event) int {
		return a.date - b.date
	})

	var currGuestCount, i int
	res := []Load{}
	for ; i < 2*n; i++ {
		for ; i < 2*n-1 && events[i].date == events[i+1].date; i++ {
			currGuestCount += events[i].delta
		}

		currGuestCount += events[i].delta
		if len(res) == 0 || res[len(res)-1].GuestCount != currGuestCount {
			res = append(res, Load{StartDate: events[i].date, GuestCount: currGuestCount})
		}
	}

	return res
}
