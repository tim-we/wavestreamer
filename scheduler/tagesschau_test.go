package scheduler

import (
	"testing"
	"time"
)

func TestExtractLatestMP3URL(t *testing.T) {
	xml := `
		<rss>
			<channel>
				<item>
					<title>Episode One</title>
					<enclosure url="https://example.com/ep1.mp3" type="audio/mpeg" length="12345"/>
					<pubDate>Mon, 22 Apr 2024 10:00:00 +0000</pubDate>
					<dc:date>2025-04-22T10:00:00Z</dc:date>
				</item>
				<item>
					<title>Episode Two</title>
					<enclosure url="https://example.com/ep2.mp3" type="audio/mpeg" length="123456"/>
					<pubDate>Tue, 23 Apr 2024 12:00:00 +0000</pubDate>
					<dc:date>2025-04-23T12:00:00Z</dc:date>
				</item>
			</channel>
		</rss>
	`

	result, err := extractLatestMP3URL([]byte(xml))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expectedURL := "https://example.com/ep2.mp3"
	expectedTime := time.Date(2024, 4, 23, 12, 0, 0, 0, time.UTC)

	if result.URL != expectedURL {
		t.Errorf("Expected URL %q, got %q", expectedURL, result.URL)
	}

	if !result.PubDate.Equal(expectedTime) {
		t.Errorf("Expected time %v, got %v", expectedTime, result.PubDate)
	}
}
