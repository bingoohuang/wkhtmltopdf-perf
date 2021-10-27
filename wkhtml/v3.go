package wkhtml

import (
	"fmt"
	"sync"
)

var (
	v3lock sync.Mutex
	v3once sync.Once
)

// ToPdfV3 use cgo bindings.
// https://forum.dlang.org/post/rgihhttgnigenuaniiql@forum.dlang.org
// On Thursday, 31 May 2018 at 19:26:12 UTC, Dr.No wrote:
// > My application create some HTML which is then converted to PDF by wkhtmltopdf library.
// > I'm trying to figure out how make the PDF generation run parallel, currently, it's running linearly.
//
// It looks like wkhtmltopdf does not support multi-threaded use; see here:
//
// https://github.com/wkhtmltopdf/wkhtmltopdf/issues/1711
//
// So, if you want to run the conversions in parallel, you will have to use separate processes.
func (*ToX) ToPdfV3(htmlURL, extraArgs string, saveFile bool) (pdfData []byte, err error) {
	v3once.Do(func() {
		//if err := pdf.Init(); err != nil {
		//	panic(err)
		//}
		//defer pdf.Destroy()
	})

	v3lock.Lock()
	defer v3lock.Unlock()
	//
	//// Create object from URL.
	//pdfObj, err := pdf.NewObject(htmlURL)
	//if err != nil {
	//	return nil, err
	//}
	//defer pdfObj.Destroy()
	//
	//// pdfObj.Footer.ContentLeft = "[date]"
	//// pdfObj.Footer.ContentCenter = "Sample footer information"
	//// pdfObj.Footer.ContentRight = "[page]"
	//// pdfObj.Footer.DisplaySeparator = true
	//
	//// Create converter.
	//converter, err := pdf.NewConverter()
	//if err != nil {
	//	return nil, err
	//}
	//defer converter.Destroy()
	//
	//// Add created objects to the converter.
	//converter.Add(pdfObj)
	//
	//// Set converter options.
	//// converter.Title = "Sample document"
	//// converter.PaperSize = pdf.A4
	//// converter.Orientation = pdf.Landscape
	//// converter.MarginTop = "1cm"
	//// converter.MarginBottom = "1cm"
	//// converter.MarginLeft = "10mm"
	//// converter.MarginRight = "10mm"
	//
	//var w io.Writer
	//var out string
	//var b bytes.Buffer
	//
	//if saveFile {
	//	if out, err = util.TempFile(".pdf"); err != nil {
	//		return nil, err
	//	}
	//	if w, err = os.Create(out); err != nil {
	//		return nil, err
	//	}
	//} else {
	//	w = &b
	//}
	//
	//// Convert objects and save the output PDF document.
	//if err := converter.Run(w); err != nil {
	//	return nil, err
	//}
	//
	//if saveFile {
	//	return []byte(out), nil
	//} else {
	//	return b.Bytes(), nil
	//}

	return nil, fmt.Errorf("Not supported")
}
