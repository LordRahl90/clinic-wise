package notes

import (
	"testing"

	"clinic-wise/pkg/testhelper"

	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/require"
)

func TestService_Feedback_ComposesInSequenceOrder(t *testing.T) {

	feedBackMap = make(map[ulid.ULID][]NoteFeeback)
	appointment := testhelper.CreateAppointment(t, db)

	svc := New(db, NewNoopWriter())

	require.NoError(t, svc.Feedback(t.Context(), NoteFeeback{
		AppointmentID: appointment.ID,
		Sequence:      2,
		Text:          "world",
		IsFinal:       false,
	}))
	require.NoError(t, svc.Feedback(t.Context(), NoteFeeback{
		AppointmentID: appointment.ID,
		Sequence:      1,
		Text:          "hello",
		IsFinal:       false,
	}))

	err := svc.Feedback(t.Context(), NoteFeeback{
		AppointmentID: appointment.ID,
		Sequence:      3,
		Text:          "ignored",
		IsFinal:       true,
	})
	require.NoError(t, err)

	res, err := svc.GetAppointmentNotes(t.Context(), appointment.DoctorID, appointment.ID)
	require.NoError(t, err)
	require.Len(t, res, 1)

	result := res[0]
	require.NotEmpty(t, result.Content)
	require.Contains(t, result.Content, "hello world ")
	require.Contains(t, result.Content, "ignored")
	require.Equal(t, "hello world ignored ", result.Content)

	//require.NoError(t, w.Close())
	//os.Stdout = stdout
	//
	//var out bytes.Buffer
	//_, err = io.Copy(&out, r)
	//require.NoError(t, err)
	//
	//printed := out.String()
	//require.Contains(t, printed, appointment.ID.String())
	//require.Contains(t, printed, "hello world ")
	//require.False(t, strings.Contains(printed, "ignored"))
}
