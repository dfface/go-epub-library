// SPDX-License-Identifier: Apache-2.0

package epub

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"
)

// Encode writes a normalized Document into EPUB ZIP stream.
func Encode(w io.Writer, doc *Document) error {
	if doc == nil {
		return fmt.Errorf("document is nil")
	}
	zw := zip.NewWriter(w)

	if err := writeMimetype(zw); err != nil {
		return err
	}
	if err := writeContainer(zw); err != nil {
		return err
	}
	if err := writePackageAndAssets(zw, doc); err != nil {
		return err
	}
	return zw.Close()
}

func writeMimetype(zw *zip.Writer) error {
	h := &zip.FileHeader{Name: "mimetype", Method: zip.Store}
	w, err := zw.CreateHeader(h)
	if err != nil {
		return err
	}
	_, err = io.WriteString(w, mimeTypeValue)
	return err
}

func writeContainer(zw *zip.Writer) error {
	w, err := zw.Create("META-INF/container.xml")
	if err != nil {
		return err
	}
	container := `<?xml version="1.0" encoding="UTF-8"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="item/standard.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`
	_, err = io.WriteString(w, container)
	return err
}

func writePackageAndAssets(zw *zip.Writer, doc *Document) error {
	assetPaths := make([]string, 0, len(doc.Assets))
	for p := range doc.Assets {
		assetPaths = append(assetPaths, p)
	}
	sort.Strings(assetPaths)

	manifestItems := make([]string, 0, len(assetPaths))
	opfDir := "item"
	for _, p := range assetPaths {
		a := doc.Assets[p]
		href := p
		if rel, err := filepath.Rel(opfDir, p); err == nil {
			href = filepath.ToSlash(rel)
		}
		manifestItems = append(manifestItems, fmt.Sprintf(`<item id=%q href=%q media-type=%q/>`, xmlEscape(a.ID), xmlEscape(href), xmlEscape(a.MimeType)))
	}

	spineItems := make([]string, 0, len(doc.Pages))
	for _, pg := range doc.Pages {
		prop := spreadToProperty(pg.Spread)
		if prop == "" {
			spineItems = append(spineItems, fmt.Sprintf(`<itemref idref=%q/>`, xmlEscape(pg.AssetID)))
			continue
		}
		spineItems = append(spineItems, fmt.Sprintf(`<itemref idref=%q properties=%q/>`, xmlEscape(pg.AssetID), xmlEscape(prop)))
	}

	rLayout := "reflowable"
	if doc.Layout == LayoutPrePaginated {
		rLayout = "pre-paginated"
	}

	opf := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<package version="3.0" xmlns="http://www.idpf.org/2007/opf" unique-identifier="pub-id">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>%s</dc:title>
    <meta property="rendition:layout">%s</meta>
  </metadata>
  <manifest>
    %s
  </manifest>
  <spine page-progression-direction=%q>
    %s
  </spine>
</package>`, xmlEscape(doc.Title), rLayout, strings.Join(manifestItems, "\n    "), xmlEscape(normalizeDirection(doc.Direction)), strings.Join(spineItems, "\n    "))

	w, err := zw.Create("item/standard.opf")
	if err != nil {
		return err
	}
	if _, err := io.WriteString(w, opf); err != nil {
		return err
	}

	for _, p := range assetPaths {
		a := doc.Assets[p]
		if a == nil || a.Open == nil {
			return fmt.Errorf("asset %s has no Open function", p)
		}
		aw, err := zw.Create(p)
		if err != nil {
			return err
		}
		rc, err := a.Open()
		if err != nil {
			return err
		}
		if _, err := io.Copy(aw, rc); err != nil {
			rc.Close()
			return err
		}
		if err := rc.Close(); err != nil {
			return err
		}
	}
	return nil
}

func spreadToProperty(spread string) string {
	switch spread {
	case "left":
		return "page-spread-left"
	case "right":
		return "page-spread-right"
	case "center":
		return "page-spread-center"
	default:
		return ""
	}
}

func xmlEscape(v string) string {
	var b bytes.Buffer
	xml.EscapeText(&b, []byte(v))
	return b.String()
}
