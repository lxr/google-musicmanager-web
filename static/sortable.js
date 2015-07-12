/*
Copyright (C) 2013 Adam Schwartz, http://adamschwartz.co

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
(function() {
  var SELECTOR, addEventListener, clickEvent, numberRegExp, sortable, touchDevice, trimRegExp;

  SELECTOR = 'table[data-sortable]';

  numberRegExp = /^-?[£$¤]?[\d,.]+%?$/;

  trimRegExp = /^\s+|\s+$/g;

  touchDevice = 'ontouchstart' in document.documentElement;

  clickEvent = touchDevice ? 'touchstart' : 'click';

  addEventListener = function(el, event, handler) {
    if (el.addEventListener != null) {
      return el.addEventListener(event, handler, false);
    } else {
      return el.attachEvent("on" + event, handler);
    }
  };

  sortable = {
    init: function(options) {
      var table, tables, _i, _len, _results;
      if (options == null) {
        options = {};
      }
      if (options.selector == null) {
        options.selector = SELECTOR;
      }
      tables = document.querySelectorAll(options.selector);
      _results = [];
      for (_i = 0, _len = tables.length; _i < _len; _i++) {
        table = tables[_i];
        _results.push(sortable.initTable(table));
      }
      return _results;
    },
    initTable: function(table) {
      var i, th, ths, _i, _len, _ref;
      if (((_ref = table.tHead) != null ? _ref.rows.length : void 0) !== 1) {
        return;
      }
      if (table.getAttribute('data-sortable-initialized') === 'true') {
        return;
      }
      table.setAttribute('data-sortable-initialized', 'true');
      ths = table.querySelectorAll('th');
      for (i = _i = 0, _len = ths.length; _i < _len; i = ++_i) {
        th = ths[i];
        if (th.getAttribute('data-sortable') !== 'false') {
          sortable.setupClickableTH(table, th, i);
        }
      }
      return table;
    },
    setupClickableTH: function(table, th, i) {
      var type;
      type = sortable.getColumnType(table, i);
      return addEventListener(th, clickEvent, function(e) {
        var newSortedDirection, row, rowArray, rowArrayObject, sorted, sortedDirection, tBody, ths, _i, _j, _k, _len, _len1, _len2, _ref, _results;
        sorted = this.getAttribute('data-sorted') === 'true';
        sortedDirection = this.getAttribute('data-sorted-direction');
        if (sorted) {
          newSortedDirection = sortedDirection === 'ascending' ? 'descending' : 'ascending';
        } else {
          newSortedDirection = type.defaultSortDirection;
        }
        ths = this.parentNode.querySelectorAll('th');
        for (_i = 0, _len = ths.length; _i < _len; _i++) {
          th = ths[_i];
          th.setAttribute('data-sorted', 'false');
          th.removeAttribute('data-sorted-direction');
        }
        this.setAttribute('data-sorted', 'true');
        this.setAttribute('data-sorted-direction', newSortedDirection);
        tBody = table.tBodies[0];
        rowArray = [];
        _ref = tBody.rows;
        for (_j = 0, _len1 = _ref.length; _j < _len1; _j++) {
          row = _ref[_j];
          rowArray.push([sortable.getNodeValue(row.cells[i]), row]);
        }
        if (sorted) {
          rowArray.reverse();
        } else {
          rowArray.sort(type.compare);
        }
        _results = [];
        for (_k = 0, _len2 = rowArray.length; _k < _len2; _k++) {
          rowArrayObject = rowArray[_k];
          _results.push(tBody.appendChild(rowArrayObject[1]));
        }
        return _results;
      });
    },
    getColumnType: function(table, i) {
      var row, text, _i, _len, _ref;
      _ref = table.tBodies[0].rows;
      for (_i = 0, _len = _ref.length; _i < _len; _i++) {
        row = _ref[_i];
        text = sortable.getNodeValue(row.cells[i]);
        if (text !== '') {
          if (text.match(numberRegExp)) {
            return sortable.types.numeric;
          }
          if (!isNaN(Date.parse(text))) {
            return sortable.types.date;
          }
        }
      }
      return sortable.types.alpha;
    },
    getNodeValue: function(node) {
      if (!node) {
        return '';
      }
      if (node.getAttribute('data-value') !== null) {
        return node.getAttribute('data-value');
      }
      if (typeof node.innerText !== 'undefined') {
        return node.innerText.replace(trimRegExp, '');
      }
      return node.textContent.replace(trimRegExp, '');
    },
    types: {
      numeric: {
        defaultSortDirection: 'descending',
        compare: function(a, b) {
          var aa, bb;
          aa = parseFloat(a[0].replace(/[^0-9.-]/g, ''), 10);
          bb = parseFloat(b[0].replace(/[^0-9.-]/g, ''), 10);
          if (isNaN(aa)) {
            aa = 0;
          }
          if (isNaN(bb)) {
            bb = 0;
          }
          return bb - aa;
        }
      },
      alpha: {
        defaultSortDirection: 'ascending',
        compare: function(a, b) {
          return a[0].localeCompare(b[0]);
        }
      },
      date: {
        defaultSortDirection: 'ascending',
        compare: function(a, b) {
          var aa, bb;
          aa = Date.parse(a[0]);
          bb = Date.parse(b[0]);
          if (isNaN(aa)) {
            aa = 0;
          }
          if (isNaN(bb)) {
            bb = 0;
          }
          return aa - bb;
        }
      }
    }
  };

  setTimeout(sortable.init, 0);

  window.Sortable = sortable;

}).call(this);