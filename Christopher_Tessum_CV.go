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

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/nickng/bibtex"
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
		Name: "Peer-Reviewed Publications <small>(*=corresponding author; self and advisees are underlined)</small>",
		Citations: []template.HTML{
			"Goodkind2025", "guo2024uncertainty", "yang2024atmospheric", "park2024",
			"giang2024", "Peshin_2024", "Schollaert_2024", "Schollaert2023", "ywang2023", "Nawaz2023",
			"gallagher2023", "jackson2023city", "thind2022environmental",
			"yuzhou2022ej", "kleiman2022", "mtessum2022",
			"thakrar2022global", "develyn2022wildfire", "wu2021reduced",
			"Balasubramanian2021", "DomingoAg2021", "TessumEJ2021", "KelpNN2020",
			"Thakrar2020", "ThindEGU2019",
			"Dimanchev2019", "Gilmore2019", "GoodkindISRM2019", "HillCorn2019", "TessumEIO2019", "LiuTrans2018",
			"PaolellaGrid2018", "Thakrar2017", "Chang2017", "Tessum2017a", "Keeler2016", "Touchaei2016",
			"Tessum2015a", "Tessum2014a", "Hu2014a", "Tessum2012", "Millet2012",
		},
	},
	{
		Name: "Preprints and Manuscripts Submitted for Review <small>(*=corresponding author; self and advisees are underlined)</small>",
		Citations: []template.HTML{
			"kazemi2024aidovecl", "KelpNN2018",
		},
	},
	// {
	// 	Name: "Manuscripts in Preparation <small>(*=corresponding author)</small>",
	// 	Citations: []template.HTML{
	// 		"ChamblissiF2018", "MullerPolicy2018",
	// 		"ThakrarInMAP2018",
	// 	},
	// },
	{
		Name:      "Reports and Other Publications",
		Citations: []template.HTML{"SABBenMAP2024", "Tessum2010a", "Tessum2010"},
	},
	{
		Name:      "Conference Papers",
		Citations: []template.HTML{"park2022learned", "koloutsou2022cloud"},
	},
	{
		Name: "Invited Presentations",
		Citations: []template.HTML{
			"Tessum2023AGU", "Tessum2023UVic", "Tessum2023NASA614",
			"Tessum2022HAQAST", "Tessum2022NASAGMAO",
			"Tessum2022AMS", "Tessum2021AGU_EJ",
			"Tessum2021UW", "Tessum2021EPRI", "Tessum2021EPAWebinar",
			"Tessum2021CACHET", "Tessum2021LBNLEAEI", "Tessum2021C40Webinar",
			"Tessum2021AMS", "Tessum2020ACM",
			"Tessum2017EIC", "Tessum2017CRC", "TessumHR2015", "TessumMEHA2015", "Tessum2014LBNL",
			"Tessum2014NatCap", "TessumBeiDa2013", "TessumChinaCDC2013", "Tessum2013AWMA",
			"TessumSETAC2012", "TessumPeking2011", "TessumMAS2011",
		},
	},
	{
		Name: "Conference Presentations",
		Citations: []template.HTML{
			"yang2023agu", "swang2023agu", "ran2023agu",
			"park2023agu", "jliu2023agu", "fatima2023agu", "guo2023agu", "park2023iama",
			"Park2023AMS", "gallagher2022scaling", "singh2022distributional", "Park2022NeurIPS", "Guo2022ACM",
			"Yang2022ACM", "Wang2022Tweeds", "wang2022addressing", "Tessum2022ISES", "Tessum2022SIAM", "Xiaokai2021IAMA",
			"Tessum2021AGU_SrcAppt", "Shiyuan2021AGU",
			"tessum2020predicting", "anenberg2020recent", "stylianou2019spatially", "kelp2019deep",
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
				Name: "Developed and taught 'CEE 492: Data Science for Civil and Environmental Engineering'",
				Time: "Fall 2020–Present",
			},
			{
				Name: "Taught 'CEE 202: Engineering Risk and Uncertainty'",
				Time: "Spring 2021–Present",
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
				Time: "2008–2023",
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
				Name: "Developer of the InMAP air quality model (https://inmap.run), which has been downloaded 5,000 times and has a user forum with 162 members",
				Time: "2012–present",
			},
			{
				Name: "Member of US EPA Science Advisory Committee for 'Review of Air Pollution Benefits Methods and Environmental Benefits Mapping and Analysis Program (BenMAP) Tool'",
				Time: "2023",
			},
			{
				Name: "Facilitator for the UIUC Grainger College of Engineering 2023 summer workshop series on 'Incorporating Computing into Engineering Curriculum'",
				Time: "2023",
			},
			{
				Name: "Member of Health Effects Institute (HEI) panel of experts to plan a Request for Proposals about electrification of diesel truck and bus fleets in the US",
				Time: "2023",
			},
			{
				Name: "Member of <i>GeoHealth</i> Early Career Editorial Board",
				Time: "2024–present",
			},
		},
	},
}

var cv2Page = []Section{
	cv[0],
	cv[1],
	{
		Name: "Selected Peer-Reviewed Publications <small>(*=corresponding author)</small>",
		Citations: []template.HTML{
			"wu2021reduced",
			"Balasubramanian2021", "DomingoAg2021",
			"TessumEJ2021", "KelpNN2020",
			"Thakrar2020", "ThindEGU2019",
			"Dimanchev2019", "GoodkindISRM2019", "HillCorn2019", "TessumEIO2019", "LiuTrans2018",
			"PaolellaGrid2018", "Tessum2017a",
			"Tessum2015a", "Tessum2014a", "Hu2014a", "Tessum2012", "Millet2012",
		},
	},
	//cv[3],
	func() Section {
		x := cv[9]
		x.Name = "Scientific, Technical, and Management Experience"
		x.Items = []Item{x.Items[0], x.Items[3]}
		return x
	}(),
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
				<i>Nature</i>,
				<i>Science</i>,
				<i>Proceedings of the National Academy of Sciences of the USA</i>,
				<i>Nature Sustainability</i>,
				<i>Nature Communications</i>,
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

func parseBibtex(bibs []string) map[template.HTML]*bibtex.BibEntry {
	out := make(map[template.HTML]*bibtex.BibEntry)
	for _, bib := range bibs {
		f, err := os.Open(bib)
		check(err)
		b := new(bytes.Buffer)
		_, err = io.Copy(b, f)
		check(err)
		elems, err := bibtex.Parse(b)
		check(err)
		for _, e := range elems.Entries {
			if _, ok := out[template.HTML(e.CiteName)]; ok {
				panic(e.CiteName)
			}
			out[template.HTML(e.CiteName)] = e
		}
	}
	return out
}

func underlineName(s string) string {
	for _, name := range []string{"Tessum, C.W.", "C.W. Tessum", "Park, M", "M. Park",
		"X. Yang", "Yang, X", "S. Wang", "Wang, S", "L. Guo", "Guo, L", "Q. Fatima", "Fatima, Q",
		"Liu, J", "J. Liu", "X. Ran", "Ran, X", "Kazemi, A", "A. Kazemi"} {
		s = strings.Replace(s, name, fmt.Sprintf("<u>%s</u>", name), -1)
	}
	return s
}

func formatCitationFunc(citations map[template.HTML]*bibtex.BibEntry) func(template.HTML) (template.HTML, error) {
	return func(key template.HTML) (template.HTML, error) {
		elem, ok := citations[key]
		if !ok {
			return "", fmt.Errorf("invalid citation key %s", key)
		}
		switch elem.Type {
		case "article":
			return template.HTML(underlineName(parseArticle(elem))), nil
		case "inproceedings":
			return template.HTML(underlineName(parseProceedings(elem))), nil
		case "techreport":
			return template.HTML(underlineName(parseReport(elem))), nil
		case "incollection":
			return template.HTML(underlineName(parseCollection(elem))), nil
		default:
			return "", fmt.Errorf("invalid citation type %s", elem.Type)
		}
	}
}

var matchDots *regexp.Regexp

func init() {
	matchDots = regexp.MustCompile(`[\.]{2,}`)
}

func parseArticle(elem *bibtex.BibEntry) string {
	for k, v := range elem.Fields {
		elem.Fields[strings.ToLower(k)] = v
	}
	title := parseTitle(elem.Fields["title"].String())
	authors := parseAuthors(elem.Fields["author"].String())
	year := parseYear(elem.Fields["year"].String())
	journal := parsePublication(elem.Fields["journal"].String())
	volume := ""
	if elem.Fields["volume"] != nil {
		volume = parseVolume(elem.Fields["volume"].String())
	}
	issue := ""
	if elem.Fields["number"] != nil {
		issue = parseIssue(elem.Fields["number"].String())
	}
	pages := ""
	if elem.Fields["pages"] != nil {
		pages = parsePages(elem.Fields["pages"].String())
	}
	url := parseURL(elem.Fields["url"].String())
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

func parseProceedings(elem *bibtex.BibEntry) string {
	title := parseTitle(elem.Fields["title"].String())
	authors := parseAuthors(elem.Fields["author"].String())
	year := parseYear(elem.Fields["year"].String())
	institution := parseBookTitle(elem.Fields["booktitle"].String())
	location := parseLocation(elem.Fields["address"].String())
	s := fmt.Sprintf("%s (%s) %s. Presented at %s, %s.", authors, year, title, institution, location)
	return s
}

func parseReport(elem *bibtex.BibEntry) string {
	title := parseTitle(elem.Fields["title"].String())
	authors := parseAuthors(elem.Fields["author"].String())
	year := parseYear(elem.Fields["year"].String())
	institution := parseBookTitle(elem.Fields["institution"].String())
	location := parseLocation(elem.Fields["address"].String())
	pages := ""
	if elem.Fields["pages"] != nil {
		pages = parsePages(elem.Fields["pages"].String())
	}
	url := ""
	if elem.Fields["url"] != nil {
		url = parseURL(elem.Fields["url"].String())
	}
	s := fmt.Sprintf("%s (%s) \"%s\", tech. rep.: %s, %s", authors, year, title, institution, location)
	if pages != "" {
		if url != "" {
			s += fmt.Sprintf(", <a href=%s>%s</a>.", url, pages)
		} else {
			s += ", " + pages + "."
		}
	} else if url != "" {
		s += fmt.Sprintf(", <a href=%s>%s</a>.", url, url)
	} else {
		s += "."
	}
	return s
}

func parseCollection(elem *bibtex.BibEntry) string {
	title := parseTitle(elem.Fields["title"].String())
	authors := parseAuthors(elem.Fields["author"].String())
	year := parseYear(elem.Fields["year"].String())
	book := parseBookTitle(elem.Fields["booktitle"].String())
	eds := removeBrackets(elem.Fields["editor"].String())
	pub := removeBrackets(elem.Fields["publisher"].String())
	pages := parsePages(elem.Fields["pages"].String())
	url := parseURL(elem.Fields["url"].String())
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
	a = strings.Replace(a, "*", "", -1)
	a = strings.TrimLeft(strings.TrimRight(a, "}"), "{")
	names := strings.Split(a, " ")
	if !strings.Contains(names[0], ",") {
		nn := len(names) - 1
		names = append(names[nn:nn+1], names[0:nn]...)
	}
	family := strings.TrimRight(strings.TrimSpace(names[0]), ",")
	given := strings.ToUpper(string(names[1][0])) + "."
	given = strings.Replace(given, "~", " ", -1)
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
