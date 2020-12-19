package main

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"

	"github.com/caltechlibrary/bibtex"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

var bibs = []string{
	"cv.bib",
	"Posters.bib",
	"Presentations.bib",
	"inprep.bib",
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
				Name:        "Assistant Professor—University of Illinois at Urbana-Champaign",
				Time:        "2020–present",
				Description: "Department of Civil and Environmental Engineering",
			},
			{
				Name:        "Research Scientist—University of Washington",
				Time:        "2016–2019",
				Description: "Department of Civil and Environmental Engineering",
			},
			{
				Name:        "Postdoctoral Associate—University of Minnesota",
				Time:        "2015–2016",
				Description: "Department of Bioproducts and Biosystems Engineering",
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
		Name: "Peer-Reviewed Publications <small>(*=corresponding author)</small>",
		Citations: []template.HTML{
			"KelpNN2020",
			"Thakrar2020", "ThindEGU2019",
			"Dimanchev2019", "Gilmore2019", "GoodkindISRM2019", "HillCorn2019", "TessumEIO2019", "LiuTrans2018",
			"PaolellaGrid2018", "Thakrar2017", "Chang2017", "Tessum2017a", "Keeler2016", "Touchaei2016",
			"Tessum2015a", "Tessum2014a", "Hu2014a", "Tessum2012", "Millet2012",
		},
	},
	{
		Name: "Preprints and Manuscripts Submitted for Review <small>(*=corresponding author)</small>",
		Citations: []template.HTML{
			"KelpNN2018",
		},
	},
	{
		Name: "Manuscripts in Preparation <small>(*=corresponding author)</small>",
		Citations: []template.HTML{
			"DomingoAg2019",
			"PaolellaEJ2018",
			"ChamblissiF2018", "MullerPolicy2018",
			"ThakrarInMAP2018",
		},
	},
	{
		Name:      "Reports and Other Publications",
		Citations: []template.HTML{"Tessum2010a", "Tessum2010"},
	},
	{
		Name: "Invited Presentations",
		Citations: []template.HTML{
			"Tessum2021AMS", "Tessum2020ACM",
			"Tessum2017EIC", "Tessum2017CRC", "TessumHR2015", "TessumMEHA2015", "Tessum2014LBNL",
			"Tessum2014NatCap", "TessumBeiDa2013", "TessumChinaCDC2013", "Tessum2013AWMA",
			"TessumSETAC2012", "TessumPeking2011", "TessumMAS2011",
		},
	},
	{
		Name: "Conference Presentations",
		Citations: []template.HTML{
			"Tessum2018CMASEIEIO", "Tessum2018CMASInMAP", "Tessum2018CMASNN",
			"Tessum2018ISEE", "Tessum2016Cobenefits", "Tessum2016ISEEa",
			"Tessum2016ISEEb", "Marshall2016HEI", "TessumAAAR2015",
			"TessumMSI2015", "Tessum2014AAAR", "Tessum2014ISEE",
			"Tessum2013ISEE", "Tessum2013MSI",
			"TessumE32011", "TessumIonE2011", "TessumLCA2011", "TessumISEE2011", "TessumMSI2011",
			"TessumE32010", "TessumBrazil2009",
		},
	},
	{
		Name: "Teaching Experience",
		Items: []Item{
			{
				Name: "Developed and taught 'Data Science for Civil and Environmental Engineering'",
				Time: "Fall 2020",
			},
			{
				Name: "Guest lectures in life cycle assessment, air pollution, and health to undergraduate and graduate students",
				Time: "2015–Present",
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
				Name: "Owner/Partner: CT Consulting LLC, Enviromind LLC",
				Time: "2008–Present",
			},
			{
				Name: "Energy Auditor: Energy Management Solutions, Inc.",
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
	/*{
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
	},*/
	{
		Name: "Synergistic Activities",
		Items: []Item{
			{
				Name: "Developed the InMAP air quality model (https://inmap.run), which has been downloaded 1,400 times (2012–present)",
			},
			{
				Name: "Grant Application Reviewer: NSF, Health Effects Institute, and US EPA",
			},
			{
				Name: "Report Peer-Reviewer: US Department of Energy (2017)",
			},
			{
				Name: "Journal Peer-Reviewer: <i>Proceedings of the National Academy of Sciences</i>, <i>Nature Sustainability</i>, <i>Environmental Science and Technology</i>, <i>Atmospheric Environment</i>,  <i>Environmental Research Letters</i>, <i>Proceedings of the Royal Society of London A</i>, <i>GeoHealth</i>, <i>Journal of Advances in Modeling Earth Systems</i>",
			},
			{
				Name: "Member: American Geophysical Union (AGU) and Association of Environmental Engineering and Science Professors (AEESP)",
			},
		},
	},
}

var cv2Page = []Section{
	cv[0],
	cv[1],
	cv[10],
	cv[2],
	cv[3],
}

var resume = []Section{
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
				Description: "Department of Bioproducts and Biosystems Engineering",
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
		Name: "Selected Publications <span style='font-variant:normal !important'><small>(*=corresponding author; full list at <a href=https://bit.ly/2DzkZoO>https://bit.ly/2DzkZoO</a>)</small></span>",
		Citations: []template.HTML{
			"KelpNN2018", "Tessum2017a", "Tessum2014a",
		},
	},
	{
		Name: "Programming Languages <span style='font-variant:normal !important'><small>(In order of experience)</small></span>",
		Items: []Item{
			{
				Name: "Go (Golang); Python; R; Javascript; SQL; FORTRAN; C; MATLAB; LabVIEW",
			},
		},
	},
	{
		Name: "Libraries and Frameworks",
		Items: []Item{
			{
				Name: "Tensorflow; Kubernetes; HPC; Google Cloud Platform; Git/Github; Travis CI; PostGIS; React",
			},
		},
	},
	{
		Name: "Open-Source Projects <span style='font-variant:normal !important'><small>(<a href=https://github.com/ctessum>https://github.com/ctessum</a>)</small></span>",
		Items: []Item{
			{
				Name: "<a href=https://github.com/spatialmodel/inmap>https://github.com/spatialmodel/inmap</a>; <a href=https://github.com/gonum/plot/>https://github.com/gonum/plot/</a>",
			},
		},
	},
	{
		Name: "Other Professional Experience",
		Items: []Item{
			{
				Name: "English Teacher: Instituto Cultural Peruano Norteamericano; Chiclayo, Peru",
				Time: "2008",
			},
			{
				Name: "Engineer: Energy Management Solutions, Inc.; Minneapolis, MN",
				Time: "2007–2008",
			},
			{
				Name: "Aerodynamics Intern: Volvo Car Corporation; Gothenburg, Sweden",
				Time: "2006",
			},
			{
				Name: "Automation Intern: Voith Paper AG; Heidenheim an der Brenz, Germany",
				Time: "2006",
			},
		},
	},
	{
		Name: "Professional Service",
		Items: []Item{
			{
				Name: "Grant Application Reviewer: NSF, Health Effects Institute, and US EPA",
			},
			{
				Name: "Report Peer-Reviewer: US Department of Energy",
			},
			{
				Name: `Journal Peer-Reviewer:
				<i>Proceedings of the National Academy of Sciences of the USA</i>,
				<i>Nature Sustainability</i>,
				<i>Environmental Science and Technology</i>,
				<i>Atmospheric Environment</i>,  <i>Environmental Research Letters</i>,
				<i>Proceedings of the Royal Society of London A</i>,
				<i>International Journal of Geographical Information Science</i>,
				<i>GeoHealth</i>, <i>Journal of Advances in Modeling Earth Systems</i>`,
			},
			{
				Name: "Member: American Geophysical Union (AGU) and Association of Environmental Engineering and Science Professors (AEESP)",
			},
		},
	},
}

func main() {

	render(cv, "Christopher_Tessum_CV.pdf")

	render(cv2Page, "Christopher_Tessum_CV_2page.pdf")

	render(resume, "Christopher_Tessum_Resume.pdf")
}

func render(cv []Section, filename string) {
	citations := parseBibtex(bibs)

	tmpl, err := template.New("cv").Funcs(map[string]interface{}{
		"ref": formatCitationFunc(citations),
	}).ParseFiles("Christopher_Tessum_CV_template.html")
	check(err)

	var b bytes.Buffer
	check(err)
	check(tmpl.ExecuteTemplate(&b, "Christopher_Tessum_CV_template.html", cv))

	printPDF(b.Bytes(), filename)
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

func underlineName(s string) string {
	s = strings.Replace(s, "C.W. Tessum", "<u>C.W. Tessum</u>", -1)
	return strings.Replace(s, "Tessum, C.W.", "<u>Tessum, C.W.</u>", -1)
}

func formatCitationFunc(citations map[template.HTML]*bibtex.Element) func(template.HTML) (template.HTML, error) {
	return func(key template.HTML) (template.HTML, error) {
		elem, ok := citations[key]
		if !ok {
			return "", fmt.Errorf("invalid citation key %s", key)
		}
		switch elem.Type {
		case "article":
			return template.HTML(underlineName(parseArticle(elem))), nil
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

var matchDots *regexp.Regexp

func init() {
	matchDots = regexp.MustCompile(`[\.]{2,}`)
}

func parseArticle(elem *bibtex.Element) string {
	title := parseTitle(elem.Tags["title"])
	authors := parseAuthors(elem.Tags["author"])
	year := parseYear(elem.Tags["year"])
	journal := parsePublication(elem.Tags["journal"])
	volume := parseVolume(elem.Tags["volume"])
	issue := parseIssue(elem.Tags["number"])
	pages := parsePages(elem.Tags["pages"])
	url := parseURL(elem.Tags["url"])
	s := authors
	if year != "" {
		s = fmt.Sprintf("%s (%s)", s, year)
	} else if !(s[len(s)-1:] == "." || s[len(s)-2:] == ".*") {
		s = fmt.Sprintf("%s.", s)
	}
	if title != "" {
		s = fmt.Sprintf("%s %s.", s, title)
	}
	if journal != "" {
		s = fmt.Sprintf("%s %s.", s, journal)
	}
	if volume != "" {
		s += " " + volume
	}
	if issue != "" {
		s = fmt.Sprintf("%s:%s", s, issue)
	}
	if pages != "" {
		if url != "" {
			s += fmt.Sprintf(" <a href=%s>%s</a>.", url, pages)
		} else {
			s += " " + pages + "."
		}
	} else if url != "" {
		s += fmt.Sprintf(" <a href=%s>%s</a>.", url, url)
	} else {
		s += "."
	}
	return matchDots.ReplaceAllString(s, ".")
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
	corresponding := strings.Contains(a, "*")
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
		if corresponding {
			s += "*"
		}
		if n == 1 {
			return s
		}
		return s + ","
	}
	s := " " + given
	if len(names) == 3 {
		s += middle
	}
	s += " " + family
	if i == n-1 {
		if corresponding {
			s += "*"
		}
		return " and " + s
	}
	if corresponding {
		s += "*"
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
	u = strings.TrimSpace(strings.TrimRight(strings.TrimLeft(u, "{"), "}"))
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

func printPDF(cv []byte, filename string) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, err := w.Write(bytes.TrimSpace(cv))
		check(err)
	}))

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	pdfPrint := chromedp.ActionFunc(func(ctx context.Context) error {
		pdf := page.PrintToPDF()
		pdf = pdf.WithMarginTop(1)
		pdf = pdf.WithMarginBottom(1)
		data, _, err := pdf.Do(ctx)
		check(err)

		o, err := os.Create(filename)
		check(err)
		_, err = o.Write(data)
		check(err)
		o.Close()
		return nil
	})

	check(chromedp.Run(ctx,
		chromedp.Navigate(ts.URL),
		pdfPrint,
	))
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
