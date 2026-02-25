package hw03frequencyanalysis

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Change to true if needed.
var taskWithAsteriskIsCompleted = false

var text = `Как видите, он  спускается  по  лестнице  вслед  за  своим
	другом   Кристофером   Робином,   головой   вниз,  пересчитывая
	ступеньки собственным затылком:  бум-бум-бум.  Другого  способа
	сходить  с  лестницы  он  пока  не  знает.  Иногда ему, правда,
		кажется, что можно бы найти какой-то другой способ, если бы  он
	только   мог   на  минутку  перестать  бумкать  и  как  следует
	сосредоточиться. Но увы - сосредоточиться-то ему и некогда.
		Как бы то ни было, вот он уже спустился  и  готов  с  вами
	познакомиться.
	- Винни-Пух. Очень приятно!
		Вас,  вероятно,  удивляет, почему его так странно зовут, а
	если вы знаете английский, то вы удивитесь еще больше.
		Это необыкновенное имя подарил ему Кристофер  Робин.  Надо
	вам  сказать,  что  когда-то Кристофер Робин был знаком с одним
	лебедем на пруду, которого он звал Пухом. Для лебедя  это  было
	очень   подходящее  имя,  потому  что  если  ты  зовешь  лебедя
	громко: "Пу-ух! Пу-ух!"- а он  не  откликается,  то  ты  всегда
	можешь  сделать вид, что ты просто понарошку стрелял; а если ты
	звал его тихо, то все подумают, что ты  просто  подул  себе  на
	нос.  Лебедь  потом  куда-то делся, а имя осталось, и Кристофер
	Робин решил отдать его своему медвежонку, чтобы оно не  пропало
	зря.
		А  Винни - так звали самую лучшую, самую добрую медведицу
	в  зоологическом  саду,  которую  очень-очень  любил  Кристофер
	Робин.  А  она  очень-очень  любила  его. Ее ли назвали Винни в
	честь Пуха, или Пуха назвали в ее честь - теперь уже никто  не
	знает,  даже папа Кристофера Робина. Когда-то он знал, а теперь
	забыл.
		Словом, теперь мишку зовут Винни-Пух, и вы знаете почему.
		Иногда Винни-Пух любит вечерком во что-нибудь поиграть,  а
	иногда,  особенно  когда  папа  дома,  он больше любит тихонько
	посидеть у огня и послушать какую-нибудь интересную сказку.
		В этот вечер...`

func TestTop10(t *testing.T) {
	t.Run("no words in empty string", func(t *testing.T) {
		require.Len(t, Top10(""), 0)
	})

	t.Run("positive test", func(t *testing.T) {
		if taskWithAsteriskIsCompleted {
			expected := []string{
				"а",         // 8
				"он",        // 8
				"и",         // 6
				"ты",        // 5
				"что",       // 5
				"в",         // 4
				"его",       // 4
				"если",      // 4
				"кристофер", // 4
				"не",        // 4
			}
			require.Equal(t, expected, Top10(text))
		} else {
			expected := []string{
				"он",        // 8
				"а",         // 6
				"и",         // 6
				"ты",        // 5
				"что",       // 5
				"-",         // 4
				"Кристофер", // 4
				"если",      // 4
				"не",        // 4
				"то",        // 4
			}
			require.Equal(t, expected, Top10(text))
		}
	})
}

func TestTop10AdditionalCases(t *testing.T) {
	t.Run("1 ties are sorted lexicographically", func(t *testing.T) {
		require.Equal(t, []string{"alpha", "bravo", "charlie"}, Top10("charlie bravo alpha"))
	})

	t.Run("2 returns only ten terms when unique terms are more than ten", func(t *testing.T) {
		input := "l k j i h g f e d c b a"
		expected := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
		require.Equal(t, expected, Top10(input))
	})

	t.Run("3 returns all terms when unique terms are exactly ten", func(t *testing.T) {
		input := "j i h g f e d c b a"
		expected := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
		require.Equal(t, expected, Top10(input))
	})

	t.Run("4 tokenizes correctly with tabs and new lines", func(t *testing.T) {
		input := "foo\tbar\nfoo  baz\n\tbar"
		expected := []string{"bar", "foo", "baz"}
		require.Equal(t, expected, Top10(input))
	})

	t.Run("5 keeps punctuation inside tokens without normalization", func(t *testing.T) {
		input := "dog,cat dog...cat dogcat dog,cat"
		expected := []string{"dog,cat", "dog...cat", "dogcat"}
		require.Equal(t, expected, Top10(input))
	})

	t.Run("6 deterministic on repeated calls with NATO Phonetic Alphabet words", func(t *testing.T) {
		natoWords := strings.Join([]string{
			"alpha bravo charlie delta echo foxtrot golf hotel india juliett kilo lima mike",
			"november oscar papa quebec romeo sierra tango uniform victor whiskey x-ray yankee zulu",
		}, " ")

		input := strings.Join([]string{
			natoWords,
			"alpha alpha bravo bravo bravo zulu zulu x-ray x-ray x-ray",
		}, " ")

		baseline := Top10(input)
		require.Equal(t, baseline, Top10(input))
		require.Equal(t, baseline, Top10(input))
	})
}
