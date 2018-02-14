package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"os"
	"strings"

	"github.com/caltechlibrary/bibtex"
)

var bibs = []string{
	"../../Mendeley Desktop/cv.bib",
	"../../Mendeley Desktop/Posters.bib",
	"../../Mendeley Desktop/Presentations.bib",
}

type Section struct {
	Name      template.HTML
	Items     []Item
	Citations []template.HTML
}

type Item struct {
	Name, Time, Description template.HTML
}

var cv = []Section{
	{
		Name: "Professional Appointments",
		Items: []Item{
			{
				Name:        "Research Scientist—University of Washington",
				Time:        "2016–Present",
				Description: "Department of Civil and Environmental Engineering",
			},
			{
				Name:        "Postdoctoral Associate—University of Minnesota",
				Time:        "2015–2016",
				Description: "Department Bioproducts and Biosystems Engineering",
			},
		},
	},
	{
		Name: "Education",
		Items: []Item{
			{
				Name: "Ph.D., Civil, Environmental, and Geo- Engineering (public health minor)—University of Minnesota",
				Time: "2009–2014",
			},
			{
				Name: "B.M.E., Mechanical Engineering (<i>cum laude</i>)—University of Minnesota",
				Time: "2002–2006",
			},
		},
	},
	{
		Name:      "Peer-Reviewed Publications",
		Citations: []template.HTML{"Thakrar2017", "Chang2017", "Tessum2017a", "Keeler2016", "Touchaei2016", "Tessum2015a", "Tessum2014a", "Hu2014a", "Tessum2012", "Millet2012"},
	},
	{
		Name:      "Reports and Other Publications",
		Citations: []template.HTML{"Tessum2010a", "Tessum2010"},
	},
	{
		Name: "Invited Presentations",
		Citations: []template.HTML{"TessumHR2015", "Tessum2014LBNL", "TessumBeiDa2013", "TessumChinaCDC2013",
			"TessumPeking2011", "TessumMAS2011"},
	},
	{
		Name:      "Conference Presentations",
		Citations: []template.HTML{"Tessum2017EIC", "Tessum2017CRC", "Tessum2016Cobenefits", "Tessum2016ISEEa", "Tessum2016ISEEb", "Marshall2016HEI", "TessumAAAR2015", "TessumMEHA2015", "TessumMSI2015", "Tessum2014AAAR", "Tessum2014ISEE", "Tessum2014NatCap", "Tessum2013ISEE", "Tessum2013AWMA", "Tessum2013MSI", "TessumSETAC2012", "TessumE32011", "TessumIonE2011", "TessumLCA2011", "TessumISEE2011", "TessumMSI2011", "TessumE32010", "TessumBrazil2009"},
	},
	{
		Name: "Teaching Experience",
		Items: []Item{
			{
				Name: "Guest lectures in life cycle assessment, air pollution, and health to undergraduate and graduate students",
				Time: "2015–2017",
			},
			{
				Name: "Teaching Assistant: Civil Engineering 5561: Air Quality Engineering, University of Minnesota",
				Time: "2013",
			},
			{
				Name: "English Teacher: Instituto Cultural Peruano Norteamericano, Chiclayo, Peru",
				Time: "2008",
			},
		},
	},
	{
		Name: "Professional Experience",
		Items: []Item{
			{
				Name: "Owner: CT Consulting",
				Time: "2008–Present",
			},
			{
				Name: "Engineer: Energy Management Solutions, Inc.",
				Time: "2007–2008",
			},
			{
				Name: "Aerodynamics Intern: Volvo Car Corporation",
				Time: "2006",
			},
			{
				Name: "Automation Intern: Voith Paper AG",
				Time: "2006",
			},
		},
	},
	{
		Name: "Honors and Awards",
		Items: []Item{
			{
				Name: "Third place student poster award: American Center for Life Cycle Analysis Annual Conference",
				Time: "2011",
			},
			{
				Name: "Admission to First Annual Fulbright US–Brazil Biofuels Short Course",
				Time: "2009",
			},
			{
				Name: "National Merit Scholarship	",
				Time: "2002–2006",
			},
		},
	},
	{
		Name: "Service",
		Items: []Item{
			{
				Name: "Reviewer for grant applications to the Health Effects Institute and the US EPA",
			},
			{
				Name: "Reviewer for publications by the Department of Energy and in journals including <i>Environmental Science and Technology</i> and <i>Atmospheric Environment</i>",
			},
			{
				Name: "Member of the International Society for Environmental Epidemiology and the American Association for Aerosol Research",
			},
		},
	},
}

func main() {
	citations := parseBibtex(bibs)

	tmpl, err := template.New("cv").Funcs(map[string]interface{}{
		"ref": formatCitationFunc(citations),
	}).ParseFiles("Christopher_Tessum_CV_template.html")
	check(err)

	w, err := os.Create("Christopher_Tessum_CV.html")
	check(err)
	check(tmpl.ExecuteTemplate(w, "Christopher_Tessum_CV_template.html", cv))
	w.Close()
}

func parseBibtex(bibs []string) map[template.HTML]*bibtex.Element {
	out := make(map[template.HTML]*bibtex.Element)
	for _, bib := range bibs {
		f, err := os.Open(bib)
		check(err)
		b := new(bytes.Buffer)
		_, err = io.Copy(b, f)
		check(err)
		elems, err := bibtex.Parse(b.Bytes())
		check(err)
		for _, e := range elems {
			for _, key := range e.Keys {
				if _, ok := out[template.HTML(key)]; ok {
					panic(key)
				}
				out[template.HTML(key)] = e
			}
		}
	}
	return out
}

func formatCitationFunc(citations map[template.HTML]*bibtex.Element) func(template.HTML) (template.HTML, error) {
	return func(key template.HTML) (template.HTML, error) {
		elem, ok := citations[key]
		if !ok {
			return "", fmt.Errorf("invalid citation key %s", key)
		}
		switch elem.Type {
		case "article":
			return template.HTML(parseArticle(elem)), nil
		case "inproceedings":
			return template.HTML(parseProceedings(elem)), nil
		case "techreport":
			return template.HTML(parseReport(elem)), nil
		case "incollection":
			return template.HTML(parseCollection(elem)), nil
		default:
			return "", fmt.Errorf("invalid citation type %s", elem.Type)
		}
	}
}

func parseArticle(elem *bibtex.Element) string {
	title := parseTitle(elem.Tags["title"])
	authors := parseAuthors(elem.Tags["author"])
	year := parseYear(elem.Tags["year"])
	journal := parsePublication(elem.Tags["journal"])
	volume := parseVolume(elem.Tags["volume"])
	issue := parseIssue(elem.Tags["issue"])
	pages := parsePages(elem.Tags["pages"])
	s := fmt.Sprintf("%s (%s) %s. %s. ", authors, year, title, journal)
	if volume != "" {
		s += volume + ": "
	}
	s += issue + " " + pages + "."
	return s
}

func parseProceedings(elem *bibtex.Element) string {
	title := parseTitle(elem.Tags["title"])
	authors := parseAuthors(elem.Tags["author"])
	year := parseYear(elem.Tags["year"])
	institution := parseBookTitle(elem.Tags["booktitle"])
	location := parseLocation(elem.Tags["address"])
	s := fmt.Sprintf("%s (%s) %s. Presented at %s, %s.", authors, year, title, institution, location)
	return s
}

func parseReport(elem *bibtex.Element) string {
	title := parseTitle(elem.Tags["title"])
	authors := parseAuthors(elem.Tags["author"])
	year := parseYear(elem.Tags["year"])
	institution := parseBookTitle(elem.Tags["institution"])
	location := parseLocation(elem.Tags["address"])
	s := fmt.Sprintf("%s (%s) \"%s\", tech. rep.: %s, %s.", authors, year, title, institution, location)
	return s
}

func parseCollection(elem *bibtex.Element) string {
	title := parseTitle(elem.Tags["title"])
	authors := parseAuthors(elem.Tags["author"])
	year := parseYear(elem.Tags["year"])
	book := parseBookTitle(elem.Tags["booktitle"])
	eds := removeBrackets(elem.Tags["editor"])
	pub := removeBrackets(elem.Tags["publisher"])
	pages := parsePages(elem.Tags["pages"])
	url := parseURL(elem.Tags["url"])
	s := fmt.Sprintf("%s (%s) \"%s\", in <i>%s</i>, ed. by %s, %s, %s, %s.", authors, year, title, book, eds, pub, pages, url)
	return s
}

func removeBrackets(s string) string {
	return strings.TrimRight(strings.TrimLeft(s, "{"), "}")
}

func parseYear(y string) string {
	return strings.TrimRight(strings.TrimLeft(y, "{"), "}")
}

func parseTitle(t string) string {
	return strings.TrimRight(strings.TrimLeft(t, "{{"), "}}")
}

func parseAuthors(a string) string {
	a = strings.Replace(a, `{\"{u}}`, "ü", -1)
	as := strings.Split(a, " and ")
	var o string
	for i, a := range as {
		o += parseName(i, len(as), a)
	}
	return o
}

func parseName(i, n int, a string) string {
	a = strings.TrimLeft(strings.TrimRight(a, "}"), "{")
	names := strings.Split(a, " ")
	family := strings.TrimRight(strings.TrimSpace(names[0]), ",")
	given := strings.ToUpper(string(names[1][0])) + "."
	var middle string
	if len(names) == 3 {
		middle = strings.TrimRight(strings.TrimSpace(string(names[2][0])), ".") + "."
	}
	if i == 0 {
		s := family + ", " + given
		if len(names) == 3 {
			s += middle
		}
		return s + ","
	}
	s := " " + given
	if len(names) == 3 {
		s += middle
	}
	s += " " + family
	if i == n-1 {
		return " and " + s
	}
	return s + ","
}

func parsePublication(p string) string {
	p = strings.TrimRight(strings.TrimLeft(p, "{"), ".}")
	return "<i>" + p + "</i>"
}

func parseVolume(v string) string {
	v = strings.TrimRight(strings.TrimLeft(v, "{"), "}")
	if v != "" {
		v = "<strong>" + v + "</strong>"
	}
	return v
}

func parseIssue(i string) string {
	i = strings.TrimRight(strings.TrimLeft(i, "{"), "}")
	return i
}

func parsePages(p string) string {
	p = strings.TrimRight(strings.TrimLeft(p, "{"), "}")
	p = strings.Replace(p, "--", "–", -1)
	return p
}

func parseURL(u string) string {
	u = strings.TrimRight(strings.TrimLeft(u, "{"), "}")
	u = strings.Replace(u, `{\_}`, "_", -1)
	return u
}

func parseBookTitle(t string) string {
	t = strings.TrimRight(strings.TrimLeft(t, "{"), "}")
	return t
}

func parseLocation(t string) string {
	t = strings.TrimRight(strings.TrimLeft(t, "{"), "}")
	t = strings.Replace(t, `{\~{a}}`, "ã", -1)
	return t
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}