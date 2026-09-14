package api

import "testing"

func TestReplyBody(t *testing.T) {
	gmailQuoted := `Yes, it ships Monday.

On Sun, Sep 14, 2026 at 6:32 PM relay+e2etest1@relay.henock.me wrote:

> --- Forwarded via Relay ---
> Alias: e2etest1@relay.henock.me
> Original from: shop@example.com
> ---
>
> Hi, can you confirm shipping?
`
	// A third-round reply: every earlier hop is quoted, nested one level deeper
	// each time, and the marker appears once per forward the user received.
	deepThread := `Monday still works.

On Sun, Sep 14, 2026 at 7:10 PM relay+e2etest1@relay.henock.me wrote:

> --- Forwarded via Relay ---
> Alias: e2etest1@relay.henock.me
> Original from: shop@example.com
> ---
>
> Great, see you then.
>
> On Sun, Sep 14, 2026 at 6:45 PM e2etest1@relay.henock.me wrote:
>
> > Yes, it ships Monday.
> >
> > On Sun, Sep 14, 2026 at 6:32 PM relay+e2etest1@relay.henock.me wrote:
> >
> > > --- Forwarded via Relay ---
> > > Alias: e2etest1@relay.henock.me
> > > Original from: shop@example.com
> > > ---
> > >
> > > Hi, can you confirm shipping?
`
	tests := []struct {
		name, stripped, full, want string
	}{
		{"mailgun stripped-text wins", "Yes, it ships Monday.", gmailQuoted, "Yes, it ships Monday."},
		{"falls back to cutting at marker", "", gmailQuoted, "Yes, it ships Monday.\n\nOn Sun, Sep 14, 2026 at 6:32 PM relay+e2etest1@relay.henock.me wrote:"},
		{"plain reply untouched", "", "Just a normal reply.", "Just a normal reply."},
		{"empty stays safe", "", "", "(no body)"},

		// multi-round threads
		{"round three, stripped-text wins", "Monday still works.", deepThread, "Monday still works."},
		{
			// Documents a known limitation of the fallback: it keeps what is
			// above the marker, so a bottom-posted reply is lost. Mailgun's
			// stripped-text finds the reply wherever it sits, so this only
			// bites when stripped-text is missing. Losing the text is the safe
			// direction to fail, since the alternative leaks the thread.
			"bottom-posted reply is lost when stripped-text is missing",
			"", "> --- Forwarded via Relay ---\n> Hi, can you confirm shipping?\n\nYes, it ships Monday.",
			"(no body)",
		},
		{"bottom-posted reply survives with stripped-text", "Yes, it ships Monday.", "> --- Forwarded via Relay ---\n> Hi\n\nYes, it ships Monday.", "Yes, it ships Monday."},
		{
			"round three fallback cuts at the first marker, dropping every nested hop",
			"", deepThread,
			"Monday still works.\n\nOn Sun, Sep 14, 2026 at 7:10 PM relay+e2etest1@relay.henock.me wrote:",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := replyBody(tc.stripped, tc.full); got != tc.want {
				t.Errorf("replyBody()\n got: %q\nwant: %q", got, tc.want)
			}
		})
	}
}
